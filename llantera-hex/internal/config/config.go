package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config centraliza la configuración de la aplicación. Todos los valores provienen de
// variables de entorno con valores por defecto seguros para desarrollo local.
type Config struct {
	HTTPHost              string
	HTTPPort              string
	DatabaseURL           string
	ServerShutdownTimeout time.Duration
	DBMaxOpenConns        int
	DBMaxIdleConns        int
	DBConnMaxLifetime     time.Duration
	AuthSecret            string
	AuthTokenTTL          time.Duration
}

// Load construye la configuración final fusionando variables de entorno.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPHost:              envOrDefault("HTTP_HOST", "0.0.0.0"),
		HTTPPort:              envOrDefault("HTTP_PORT", "8081"),
		DatabaseURL:           envOrDefault("DATABASE_URL", "postgres://postgres:Daiki87.google@127.0.0.1:5432/llantera?sslmode=disable"),
		ServerShutdownTimeout: parseDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
		DBMaxOpenConns:        ParseIntEnv("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:        ParseIntEnv("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime:     parseDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		AuthSecret:            envOrDefault("AUTH_SECRET", "dev-secret-change-me"),
		AuthTokenTTL:          parseDuration("AUTH_TOKEN_TTL", 24*6*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("la variable DATABASE_URL es obligatoria")
	}

	return cfg, nil
}

// HTTPAddress devuelve host:puerto listo para net/http.
func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%s", c.HTTPHost, c.HTTPPort)
}

func envOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(key string, defaultValue time.Duration) time.Duration {
	value := envOrDefault(key, "")
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

// ParseIntEnv permite leer enteros positivos desde el entorno con default seguro.
func ParseIntEnv(key string, defaultValue int) int {
	value := envOrDefault(key, "")
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
