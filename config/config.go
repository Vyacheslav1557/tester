package config

type Config struct {
	Env string `env:"ENV" env-default:"prod"`

	Address string `env:"ADDRESS" required:"true"`

	Pandoc      string `env:"PANDOC" required:"true"`
	PostgresDSN string `env:"POSTGRES_DSN" required:"true"`
	RedisDSN    string `env:"REDIS_DSN" required:"true"`

	JWTSecret string `env:"JWT_SECRET" required:"true"`

	AdminUsername string `env:"ADMIN_USERNAME" env-default:"admin"`
	AdminPassword string `env:"ADMIN_PASSWORD" env-default:"admin"`

	S3Endpoint  string `env:"S3_ENDPOINT" required:"true"`
	S3AccessKey string `env:"S3_ACCESS_KEY" required:"true"`
	S3SecretKey string `env:"S3_SECRET_KEY" required:"true"`

	CacheDir string `env:"CACHE_DIR" env-default:"/tmp"`

	NatsUrl string `env:"NATS_URL" env-default:"nats://localhost:4222"`

	//RabbitDSN    string `env:"RABBIT_DSN" required:"true"`
	//InstanceName string `env:"INSTANCE_NAME" required:"true"`
	//RQueueName   string `env:"R_QUEUE_NAME" required:"true"`
	//TQueueName   string `env:"T_QUEUE_NAME" required:"true"`
	//NQueueName   string `env:"N_QUEUE_NAME" required:"true"`
}
