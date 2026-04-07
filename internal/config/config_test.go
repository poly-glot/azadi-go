package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Environment != "dev" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "dev")
	}
	if cfg.GCPProjectID != "demo-azadi" {
		t.Errorf("GCPProjectID = %q, want %q", cfg.GCPProjectID, "demo-azadi")
	}
	if cfg.FirestoreDB != "azadi" {
		t.Errorf("FirestoreDB = %q, want %q", cfg.FirestoreDB, "azadi")
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("GCP_PROJECT_ID", "my-project")
	t.Setenv("SEED_DATA", "true")

	cfg := Load()

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}
	if cfg.GCPProjectID != "my-project" {
		t.Errorf("GCPProjectID = %q, want %q", cfg.GCPProjectID, "my-project")
	}
	if cfg.SeedData != true {
		t.Errorf("SeedData = %v, want true", cfg.SeedData)
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"dev environment", "dev", true},
		{"production environment", "production", false},
		{"staging environment", "staging", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			if got := cfg.IsDev(); got != tt.want {
				t.Errorf("IsDev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvStr(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		want     string
	}{
		{"set value", "TEST_STR_SET", "hello", "default", "hello"},
		{"unset uses fallback", "TEST_STR_UNSET", "", "default", "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				t.Setenv(tt.key, tt.value)
			}
			if got := envStr(tt.key, tt.fallback); got != tt.want {
				t.Errorf("envStr(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback int
		want     int
	}{
		{"valid int", "TEST_INT_VALID", "42", 0, 42},
		{"invalid int uses fallback", "TEST_INT_INVALID", "abc", 99, 99},
		{"unset uses fallback", "TEST_INT_UNSET", "", 8080, 8080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				t.Setenv(tt.key, tt.value)
			}
			if got := envInt(tt.key, tt.fallback); got != tt.want {
				t.Errorf("envInt(%q, %d) = %d, want %d", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback bool
		want     bool
	}{
		{"true string", "TEST_BOOL_TRUE", "true", false, true},
		{"false string", "TEST_BOOL_FALSE", "false", true, false},
		{"1 is true", "TEST_BOOL_ONE", "1", false, true},
		{"invalid uses fallback", "TEST_BOOL_INVALID", "yes-please", false, false},
		{"unset uses fallback true", "TEST_BOOL_UNSET_T", "", true, true},
		{"unset uses fallback false", "TEST_BOOL_UNSET_F", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				t.Setenv(tt.key, tt.value)
			}
			if got := envBool(tt.key, tt.fallback); got != tt.want {
				t.Errorf("envBool(%q, %v) = %v, want %v", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}
