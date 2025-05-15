package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type Config struct {
	Port                 string      `mapstructure:"PORT"`
	Domain               string      `mapstructure:"DOMAIN"`
	Mode                 Environment `mapstructure:"MODE"`
	DBPath               string      `mapstructure:"DB_FILE_PATH"`
	CookieStoreSecretKey string      `mapstructure:"COOKIE_SECRET"`
	AdminUserID          string      `mapstructure:"ADMIN_USER_ID"`
	AdminUserPassword    string      `mapstructure:"ADMIN_USER_PASSWORD"`
	JWTSecretKey         string      `mapstructure:"JWT_SECRET_KEY"`
	StripeWebhookSecret  string      `mapstructure:"STRIPE_WEBHOOK_SECRET"`
	StripeAPIKey         string      `mapstructure:"STRIPE_API_KEY"`
}

func Load() (*Config, error) {

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("DB_FILE_PATH", "main.db")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Warning: .env file could not be found %v", err)
		} else {
			return nil, err
		}
	}

	var config Config
	var errs []error

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("viper could not unmarshal config vals from .env: %w", err)
	}

	if config.Port == "" {
		errs = append(errs, errors.New("env PORT value not set in env"))
	}
	if config.Mode != Development && config.Mode != Production {
		errs = append(errs, errors.New("env MODE not set in env"))
	}
	if config.Domain == "" {
		errs = append(errs, errors.New("env DOMAIN value not set in env"))
	}
	if config.AdminUserID == "" {
		errs = append(errs, errors.New("env ADMIN_USER_ID not set"))
	}
	if config.AdminUserPassword == "" {
		errs = append(errs, errors.New("env ADMIN_USER_PASSWORD not set"))
	}
	if config.StripeAPIKey == "" {
		errs = append(errs, errors.New("env STRIPE_API_KEY not set"))
	}
	if config.StripeWebhookSecret == "" {
		errs = append(errs, errors.New("env STRIPE_WEBHOOK_SECRET not set"))
	}
	if config.CookieStoreSecretKey == "" {
		errs = append(errs, errors.New("env COOKIE_SECRET not set"))
	}
	if config.JWTSecretKey == "" {
		errs = append(errs, errors.New("env JWT_SECRET_KEY not set"))
	}
	if len(errs) > 0 {
		// Combine errors for better reporting
		combinedErr := errors.New("configuration errors")
		for _, e := range errs {
			combinedErr = fmt.Errorf("%w\n- %s", combinedErr, e.Error())
		}
		return nil, combinedErr
	}

	return &config, nil
}
