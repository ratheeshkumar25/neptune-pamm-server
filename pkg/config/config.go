// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/config/config.go
// Role: Configuration Management
// Description: Centralized, env-driven configuration. Values are loaded from a local
// .env file when present and always overridable by real process environment variables
// (so `docker run --env-file` and orchestrator-injected env work without a file).

package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration, grouped by concern.
type Config struct {
	Server   ServerConfig   `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	Mongo    MongoConfig    `mapstructure:",squash"`
	Nats     NatsConfig     `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	Auth     AuthConfig     `mapstructure:",squash"`
	SMTP     SMTPConfig     `mapstructure:",squash"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	ServerPort int `mapstructure:"server_port"`
	UserPort   int `mapstructure:"user_port"`
	FeesPort   int `mapstructure:"fees_port"`
}

// PostgresConfig holds the transactional DB / ledger connection settings.
type PostgresConfig struct {
	DBHost     string `mapstructure:"db_host"`
	DBPort     int    `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	DBSSLMode  string `mapstructure:"db_sslmode"`
}

// MongoConfig holds the events / time-series store settings.
type MongoConfig struct {
	MongoURI string `mapstructure:"mongo_uri"`
}

// NatsConfig holds the message broker settings.
type NatsConfig struct {
	NatsURL string `mapstructure:"nats_url"`
}

// RedisConfig holds the cache / hot-state / session store settings.
type RedisConfig struct {
	RedisHost     string `mapstructure:"redis_host"`
	RedisPort     int    `mapstructure:"redis_port"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret        string `mapstructure:"jwt_secret"`
	JWTIssuer        string `mapstructure:"jwt_issuer"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLHours  int    `mapstructure:"refresh_ttl_hours"`
}

// AccessTTL is the access-token lifetime.
func (c AuthConfig) AccessTTL() time.Duration {
	return time.Duration(c.AccessTTLMinutes) * time.Minute
}

// RefreshTTL is the refresh-token lifetime.
func (c AuthConfig) RefreshTTL() time.Duration {
	return time.Duration(c.RefreshTTLHours) * time.Hour
}

// SMTPConfig holds outbound email settings.
type SMTPConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
}

// envKeys lists every key bound from the environment. Binding explicitly makes
// AutomaticEnv work even when no .env file is present (e.g. inside a container).
var envKeys = []string{
	"server_port", "user_port", "fees_port",
	"db_host", "db_port", "db_user", "db_password", "db_name", "db_sslmode",
	"mongo_uri",
	"nats_url",
	"redis_host", "redis_port", "redis_password", "redis_db",
	"jwt_secret", "jwt_issuer", "access_ttl_minutes", "refresh_ttl_hours",
	"smtp_host", "smtp_port", "smtp_user", "smtp_password",
}

// LoadConfig reads configuration from a .env file (if present) and the process
// environment, applies defaults, validates required fields, and returns a Config.
func LoadConfig(path ...string) (*Config, error) {
	v := viper.New()

	// Read a .env file when available; absence is not fatal (env may come from the host).
	v.SetConfigName(".env")
	v.SetConfigType("env")
	if len(path) > 0 && path[0] != "" {
		v.AddConfigPath(path[0])
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./..")

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	// Real environment variables take precedence over the file.
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for _, k := range envKeys {
		_ = v.BindEnv(k)
	}

	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults provides safe fallbacks for non-secret values.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server_port", 3000)
	v.SetDefault("db_host", "localhost")
	v.SetDefault("db_port", 5432)
	v.SetDefault("db_sslmode", "disable")
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", 6379)
	v.SetDefault("redis_db", 0)
	v.SetDefault("jwt_issuer", "neptune-pamm")
	v.SetDefault("access_ttl_minutes", 15)
	v.SetDefault("refresh_ttl_hours", 168) // 7 days
}

// validate ensures the fields the service cannot start without are present.
func (c *Config) validate() error {
	var missing []string
	if c.Postgres.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if c.Postgres.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if c.Auth.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}
	return nil
}

// DSN returns a PostgreSQL connection string built from the Postgres settings.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Addr returns the host:port address for Redis.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}
