package config

type Application struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`

	AdminUsername string `env:"ADMIN_USERNAME" envDefault:"Admin"`
	AdminPass     string `env:"ADMIN_PASS" envDefault:"Admin"`

	API
}
