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

	//RabbitDSN    string `env:"RABBIT_DSN" required:"true"`
	//InstanceName string `env:"INSTANCE_NAME" required:"true"`
	//RQueueName   string `env:"R_QUEUE_NAME" required:"true"`
	//TQueueName   string `env:"T_QUEUE_NAME" required:"true"`
	//NQueueName   string `env:"N_QUEUE_NAME" required:"true"`
}
