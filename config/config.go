package config

import (
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Server   *Server   `mapstructure:"server" validate:"required"`
		Database *Database `mapstructure:"database" validate:"required"`
		Env      *Env      `mapstructure:"env" validate:"required"`
	}

	Server struct {
		Port         int           `mapstructure:"port" validate:"required"`
		Environment  string        `mapstructure:"environment" validate:"required"`
		BodyLimit    int           `mapstructure:"bodyLimit" validate:"required"`
		TimeOut      time.Duration `mapstructure:"timeout" validate:"required"`
		ClientURL    string        `mapstructure:"client_url" validate:"required"`
		ServerDomain string        `mapstructure:"server_domain" validate:"required"`
	}

	Database struct {
		Driver   string `mapstructure:"driver" validate:"required"`
		Host     string `mapstructure:"host" validate:"required"`
		Port     int    `mapstructure:"port" validate:"required"`
		User     string `mapstructure:"user" validate:"required"`
		Password string `mapstructure:"password" validate:"required"`
		DBName   string `mapstructure:"dbname" validate:"required"`
		SSLMode  string `mapstructure:"sslmode" validate:"required"`
		Schema   string `mapstructure:"schema" validate:"required"`
	}

	Env struct {
		SecretKey string `mapstructure:"secretkey" validate:"required"`
	}
)

var (
	once           sync.Once
	configInstance *Config
)

func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// Allow env vars to override sensitive fields: DATABASE_PASSWORD, ENV_SECRETKEY
		viper.BindEnv("database.password", "DATABASE_PASSWORD")
		viper.BindEnv("env.secretkey", "ENV_SECRETKEY")

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(&configInstance); err != nil {
			panic(err)
		}

		validating := validator.New()

		if err := validating.Struct(configInstance); err != nil {
			panic(err)
		}
	})

	return configInstance
}
