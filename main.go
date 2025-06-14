package main

import (
	"context"
	"fmt"
	"github.com/Vyacheslav1557/tester/config"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/auth"
	authHandlers "github.com/Vyacheslav1557/tester/internal/auth/delivery/rest"
	authUseCase "github.com/Vyacheslav1557/tester/internal/auth/usecase"
	"github.com/Vyacheslav1557/tester/internal/contests"
	contestsHandlers "github.com/Vyacheslav1557/tester/internal/contests/delivery/rest"
	contestsRepository "github.com/Vyacheslav1557/tester/internal/contests/repository"
	contestsUseCase "github.com/Vyacheslav1557/tester/internal/contests/usecase"
	"github.com/Vyacheslav1557/tester/internal/middleware"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	problemsHandlers "github.com/Vyacheslav1557/tester/internal/problems/delivery/rest"
	problemsRepository "github.com/Vyacheslav1557/tester/internal/problems/repository"
	problemsUseCase "github.com/Vyacheslav1557/tester/internal/problems/usecase"
	sessionsRepository "github.com/Vyacheslav1557/tester/internal/sessions/repository"
	sessionsUseCase "github.com/Vyacheslav1557/tester/internal/sessions/usecase"
	"github.com/Vyacheslav1557/tester/internal/solutions"
	solutionsHandlers "github.com/Vyacheslav1557/tester/internal/solutions/delivery/rest"
	solutionsRepository "github.com/Vyacheslav1557/tester/internal/solutions/repository"
	solutionsUseCase "github.com/Vyacheslav1557/tester/internal/solutions/usecase"
	"github.com/Vyacheslav1557/tester/internal/users"
	usersHandlers "github.com/Vyacheslav1557/tester/internal/users/delivery/rest"
	usersRepository "github.com/Vyacheslav1557/tester/internal/users/repository"
	usersUseCase "github.com/Vyacheslav1557/tester/internal/users/usecase"
	"github.com/Vyacheslav1557/tester/pkg"
	tester "github.com/Vyacheslav1557/tester/pkg/tester"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var cfg config.Config
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		panic(fmt.Sprintf("error reading config: %s", err.Error()))
	}

	var logger *zap.Logger
	if cfg.Env == "prod" {
		logger = zap.Must(zap.NewProduction())
	} else if cfg.Env == "dev" {
		lcfg := zap.NewDevelopmentConfig()
		lcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		lcfg.Encoding = "console"
		lcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		lcfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		lcfg.DisableStacktrace = true

		logger, err = lcfg.Build()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()
	} else {
		panic(fmt.Sprintf(`error reading config: env expected "prod" or "dev", got "%s"`, cfg.Env))
	}

	logger.Info("connecting to postgres")
	db, err := pkg.NewPostgresDB(cfg.PostgresDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	logger.Info("successfully connected to postgres")

	logger.Info("connecting to redis")
	vk, err := pkg.NewValkeyClient(cfg.RedisDSN)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error connecting to redis: %v", err))
	}
	logger.Info("successfully connected to redis")

	logger.Info("connecting to s3")
	s3Client, err := pkg.NewS3Client(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error connecting to s3: %v", err))
	}
	logger.Info("successfully connected to s3")

	usersRepo := usersRepository.NewRepository(db)

	sessionsRepo := sessionsRepository.NewValkeyRepository(vk)
	sessionsUC := sessionsUseCase.NewUseCase(sessionsRepo, cfg)

	usersUC := usersUseCase.NewUseCase(sessionsRepo, usersRepo)

	// every time we start the app, we create an admin user
	// think of a better way
	_, err = usersUC.CreateUser(context.Background(),
		&models.UserCreation{
			Username: cfg.AdminUsername,
			Password: cfg.AdminPassword,
			Role:     models.RoleAdmin,
		})
	if err != nil {
		logger.Error(fmt.Sprintf("error creating admin user: %s", err.Error()))
	}

	np, err := pkg.NewNatsPublisher(cfg.NatsUrl)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error connecting to nats: %v", err))
	}

	authUC := authUseCase.NewUseCase(usersUC, sessionsUC)

	pandocClient := pkg.NewPandocClient(&http.Client{}, cfg.Pandoc)

	problemsRepo := problemsRepository.NewRepository(db)
	s3Repo := problemsRepository.NewS3Repository(s3Client, "tester-problems-archives")

	problemsUC := problemsUseCase.NewUseCase(problemsRepo, pandocClient, s3Repo, cfg.CacheDir)

	contestsRepo := contestsRepository.NewRepository(db)
	contestsUC := contestsUseCase.NewContestUseCase(contestsRepo)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(fmt.Errorf("failed to create docker client: %v", err))
	}

	t := tester.NewTester(cfg.CacheDir, tester.NewDockerExecutor(cli), 2)

	solutionsRepo := solutionsRepository.NewRepository(db)
	solutionsUC := solutionsUseCase.NewUseCase(solutionsRepo, problemsUC, np, t)

	if err := os.MkdirAll(cfg.CacheDir, 0700); err != nil {
		panic(fmt.Errorf("failed to create cache dir: %v", err))
	}

	server := fiber.New(fiber.Config{
		BodyLimit: 512 * 1024 * 1024, // 512 MB
	})

	type MergedHandlers struct {
		users.UsersHandlers
		auth.AuthHandlers
		contests.ContestsHandlers
		problems.ProblemsHandlers
		solutions.SolutionsHandlers
	}

	merged := MergedHandlers{
		usersHandlers.NewHandlers(usersUC),
		authHandlers.NewHandlers(authUC, cfg.JWTSecret),
		contestsHandlers.NewHandlers(problemsUC, contestsUC),
		problemsHandlers.NewHandlers(problemsUC),
		solutionsHandlers.NewHandlers(solutionsUC, problemsUC, contestsUC),
	}

	testerv1.RegisterHandlersWithOptions(server, merged, testerv1.FiberServerOptions{
		Middlewares: []testerv1.MiddlewareFunc{
			middleware.ErrorHandlerMiddleware(logger),
			middleware.AuthMiddleware(cfg.JWTSecret, sessionsUC),
			//cors.New(cors.Config{
			//	AllowOrigins:     "http://localhost:3000",
			//	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			//	AllowHeaders:     "Content-Type,Set-Cookie,Credentials",
			//	AllowCredentials: true,
			//}),
		},
	})

	// here we handle REST endpoint response or websocket upgrade
	//server.Use("/solutions", websocket.New(merged.ListSolutionsWS))

	go func() {
		err := server.Listen(cfg.Address)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error starting server: %s", err.Error()))
		}
	}()

	logger.Info(fmt.Sprintf("server started on %s", cfg.Address))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
}
