package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  int
	Environment           string
	GCPProjectID          string
	FirestoreEmulatorHost string
	FirestoreDB           string
	StripeAPIKey          string
	StripeWebhookSecret   string
	StripePublishableKey  string
	ResendAPIKey          string
	FromEmail             string
	EncryptionKey         string
	EncryptionSalt        string
	ViteDevURL            string
	SeedData              bool
	AppDomain             string
}

func Load() *Config {
	return &Config{
		Port:                  envInt("PORT", 8080),
		Environment:           envStr("ENVIRONMENT", "dev"),
		GCPProjectID:          envStr("GCP_PROJECT_ID", "demo-azadi"),
		FirestoreEmulatorHost: envStr("FIRESTORE_EMULATOR_HOST", ""),
		FirestoreDB:           envStr("FIRESTORE_DB", "azadi"),
		StripeAPIKey:          envStr("STRIPE_API_KEY", ""),
		StripeWebhookSecret:   envStr("STRIPE_WEBHOOK_SECRET", ""),
		StripePublishableKey:  envStr("VITE_STRIPE_PUBLISHABLE_KEY", ""),
		ResendAPIKey:          envStr("RESEND_API_KEY", ""),
		FromEmail:             envStr("RESEND_FROM_EMAIL", "noreply@junaid.guru"),
		EncryptionKey:         envStr("AZADI_ENCRYPTION_KEY", "dev-only-key-change-in-prod-32ch"),
		EncryptionSalt:        envStr("AZADI_ENCRYPTION_SALT", "a1b2c3d4e5f6a7b8"),
		ViteDevURL:            envStr("VITE_DEV_URL", ""),
		SeedData:              envBool("SEED_DATA", true),
		AppDomain:             envStr("APP_DOMAIN", "localhost:8080"),
	}
}

func (c *Config) IsDev() bool {
	return c.Environment == "dev"
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
