package config

type Config struct {
	DBConnectionString string `env:"DB_CONNECTION_STRING" envDefault:"postgres://user:password@localhost:5432/corroboros?sslmode=disable"`
	MaxUploadSize      int64  `env:"MAX_UPLOAD_SIZE" envDefault:"10485760"` // 10 MB
	ServerPort         string `env:"SERVER_PORT" envDefault:"8080"`
}
