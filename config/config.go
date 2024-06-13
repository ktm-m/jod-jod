package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"strings"
	"sync"
	"time"
)

type (
	Database struct {
		Host     string `mapstructure:"host" validate:"required"`
		Port     int    `mapstructure:"port" validate:"required"`
		Username string `mapstructure:"username" validate:"required"`
		Password string `mapstructure:"password" validate:"required"`
		Database string `mapstructure:"database" validate:"required"`
		SSLMode  string `mapstructure:"sslmode" validate:"required"`
	}

	Server struct {
		Host    string        `mapstructure:"host" validate:"required"`
		Port    int           `mapstructure:"port" validate:"required"`
		Name    string        `mapstructure:"name" validate:"required"`
		Version string        `mapstructure:"version" validate:"required"`
		Timeout time.Duration `mapstructure:"timeout" validate:"required"`
	}

	Auth struct {
		Secret string `mapstructure:"secret" validate:"required"`
	}

	AWS struct {
		Region          string `mapstructure:"region" validate:"required"`
		Bucket          string `mapstructure:"bucket" validate:"required"`
		BucketSlipPath  string `mapstructure:"bucket_slip_path" validate:"required"`
		AccessKeyID     string `mapstructure:"access_key_id" validate:"required"`
		SecretAccessKey string `mapstructure:"secret_access_key" validate:"required"`
	}

	Config struct {
		Database *Database `mapstructure:"database" validate:"required"`
		Server   *Server   `mapstructure:"server" validate:"required"`
		Auth     *Auth     `mapstructure:"auth" validate:"required"`
		AWS      *AWS      `mapstructure:"aws" validate:"required"`
	}
)

var (
	once sync.Once
	cfg  *Config
)

func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			panic(err)
		}

		validate := validator.New()
		if err := validate.Struct(cfg); err != nil {
			panic(err)
		}
	})
	return cfg
}
