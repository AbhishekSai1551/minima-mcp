package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Minima    MinimaConfig
	RateLimit RateLimitConfig
	Audit     AuditConfig
	Auth      AuthConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogLevel     string
}

type MinimaConfig struct {
	DataHost      string
	DataPort      int
	MiniDAPPHost  string
	MiniDAPPPort  int
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	CleanupInterval   time.Duration
}

type AuditConfig struct {
	Enabled    bool
	LogDir     string
	MaxAgeDays int
	Format     string
}

type AuthConfig struct {
	Enabled      bool
	Mode         string
	OnboardingDN string
	JWTSecret    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 3001),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			LogLevel:     getEnv("LOG_LEVEL", "info"),
		},
		Minima: MinimaConfig{
			DataHost:      getEnv("MINIMA_DATA_HOST", "localhost"),
			DataPort:      getEnvInt("MINIMA_DATA_PORT", 9004),
			MiniDAPPHost:  getEnv("MINIMA_MINIDAPP_HOST", "localhost"),
			MiniDAPPPort:  getEnvInt("MINIMA_MINIDAPP_PORT", 9005),
			Timeout:       getEnvDuration("MINIMA_TIMEOUT", 30*time.Second),
			RetryAttempts: getEnvInt("MINIMA_RETRY_ATTEMPTS", 3),
			RetryDelay:    getEnvDuration("MINIMA_RETRY_DELAY", 1*time.Second),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvFloat("RATE_LIMIT_RPS", 10),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 20),
			CleanupInterval:   getEnvDuration("RATE_LIMIT_CLEANUP", 5*time.Minute),
		},
		Audit: AuditConfig{
			Enabled:    getEnvBool("AUDIT_ENABLED", true),
			LogDir:     getEnv("AUDIT_LOG_DIR", "./audit-logs"),
			MaxAgeDays: getEnvInt("AUDIT_MAX_AGE_DAYS", 90),
			Format:     getEnv("AUDIT_FORMAT", "json"),
		},
		Auth: AuthConfig{
			Enabled:      getEnvBool("AUTH_ENABLED", true),
			Mode:         getEnv("AUTH_MODE", "tiered"),
			OnboardingDN: getEnv("AUTH_ONBOARDING_DN", ""),
			JWTSecret:    getEnv("AUTH_JWT_SECRET", ""),
		},
	}

	if cfg.Minima.DataPort < 1 || cfg.Minima.DataPort > 65535 {
		return nil, fmt.Errorf("invalid MINIMA_DATA_PORT: %d", cfg.Minima.DataPort)
	}
	if cfg.Minima.MiniDAPPPort < 1 || cfg.Minima.MiniDAPPPort > 65535 {
		return nil, fmt.Errorf("invalid MINIMA_MINIDAPP_PORT: %d", cfg.Minima.MiniDAPPPort)
	}

	return cfg, nil
}

func (m *MinimaConfig) DataURL() string {
	return fmt.Sprintf("http://%s:%d", m.DataHost, m.DataPort)
}

func (m *MinimaConfig) MiniDAPPURL() string {
	return fmt.Sprintf("http://%s:%d", m.MiniDAPPHost, m.MiniDAPPPort)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
