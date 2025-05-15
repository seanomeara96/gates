package config

import (
	"errors"
	"fmt"
	"os"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type Config struct {
	Port                 string
	Domain               string
	Mode                 Environment
	DBPath               string
	CookieStoreSecretKey string
	AdminUserID          string
	AdminUserPassword    string
	JWTSecretKey         string
	StripeWebhookSecret  string
	StripeAPIKey         string
}

func Load() (*Config, error) {

	var config Config
	var errs []error

	config.Port = os.Getenv("PORT")
	if config.Port == "" {
		errs = append(errs, errors.New("env PORT value not set in env"))
	}

	config.Mode = Environment(os.Getenv("MODE"))
	if config.Mode != Development && config.Mode != Production {
		errs = append(errs, errors.New("env MODE not set in env"))
	}

	config.Domain = os.Getenv("DOMAIN")
	if config.Domain == "" {
		errs = append(errs, errors.New("env DOMAIN value not set in env"))
	}

	config.DBPath = os.Getenv("DB_FILE_PATH")
	if config.DBPath == "" {
		config.DBPath = "main.db"
	}

	config.AdminUserID = os.Getenv("ADMIN_USER_ID")
	if config.AdminUserID == "" {
		errs = append(errs, errors.New("env ADMIN_USER_ID not set"))
	}

	config.AdminUserPassword = os.Getenv("ADMIN_USER_PASSWORD")
	if config.AdminUserPassword == "" {
		errs = append(errs, errors.New("env ADMIN_USER_PASSWORD not set"))
	}

	config.StripeAPIKey = os.Getenv("STRIPE_API_KEY")
	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	config.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
	if config.StripeWebhookSecret == "" {
		errs = append(errs, errors.New("env STRIPE_WEBHOOK_SECRET not set"))
	}

	config.CookieStoreSecretKey = os.Getenv("COOKIE_SECRET")
	if config.CookieStoreSecretKey == "" {
		errs = append(errs, errors.New("env COOKIE_SECRET not set"))
	}

	config.JWTSecretKey = os.Getenv("JWT_SECRET_KEY")
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
