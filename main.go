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
	"github.com/Vyacheslav1557/tester/internal/users"
	usersHandlers "github.com/Vyacheslav1557/tester/internal/users/delivery/rest"
	usersRepository "github.com/Vyacheslav1557/tester/internal/users/repository"
	usersUseCase "github.com/Vyacheslav1557/tester/internal/users/usecase"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
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
		logger = zap.Must(zap.NewDevelopment())
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
		logger.Fatal(fmt.Sprintf("error connecting to redis: %s", err.Error()))
	}
	logger.Info("successfully connected to redis")

	usersRepo := usersRepository.NewRepository(db)

	_, err = usersRepo.CreateUser(context.Background(),
		usersRepo.DB(), &models.UserCreation{
			Username: cfg.AdminUsername,
			Password: cfg.AdminPassword,
			Role:     models.RoleAdmin,
		})
	if err != nil {
		logger.Error(fmt.Sprintf("error creating admin user: %s", err.Error()))
	}

	sessionsRepo := sessionsRepository.NewValkeyRepository(vk)
	sessionsUC := sessionsUseCase.NewUseCase(sessionsRepo, cfg)

	usersUC := usersUseCase.NewUseCase(sessionsRepo, usersRepo)

	authUC := authUseCase.NewUseCase(usersUC, sessionsUC)

	pandocClient := pkg.NewPandocClient(&http.Client{}, cfg.Pandoc)

	problemsRepo := problemsRepository.NewRepository(db)
	problemsUC := problemsUseCase.NewUseCase(problemsRepo, pandocClient)

	contestsRepo := contestsRepository.NewRepository(db)
	contestsUC := contestsUseCase.NewContestUseCase(contestsRepo)

	server := fiber.New()

	type MergedHandlers struct {
		users.UsersHandlers
		auth.AuthHandlers
		contests.ContestsHandlers
		problems.ProblemsHandlers
	}

	merged := MergedHandlers{
		usersHandlers.NewHandlers(usersUC),
		authHandlers.NewHandlers(authUC, cfg.JWTSecret),
		contestsHandlers.NewHandlers(problemsUC, contestsUC, cfg.JWTSecret),
		problemsHandlers.NewHandlers(problemsUC),
	}

	testerv1.RegisterHandlersWithOptions(server, merged, testerv1.FiberServerOptions{
		Middlewares: []testerv1.MiddlewareFunc{
			fiberlogger.New(),
			middleware.AuthMiddleware(cfg.JWTSecret, sessionsUC),
			//rest.AuthMiddleware(cfg.JWTSecret, userUC),
			//cors.New(cors.Config{
			//	AllowOrigins:     "http://localhost:3000",
			//	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			//	AllowHeaders:     "Content-Type,Set-Cookie,Credentials",
			//	AllowCredentials: true,
			//}),
		},
	})

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
