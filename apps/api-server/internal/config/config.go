package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server         ServerConfig         `json:"server"`
	Database       DatabaseConfig       `json:"database"`
	Redis          RedisConfig          `json:"redis"`
	IMSI           IMSIConfig           `json:"imsi"`
	ES2            ES2Config            `json:"es2"`
	Payment        PaymentConfig        `json:"payment"`
	ChargingEngine ChargingEngineConfig `json:"charging_engine"`
	InfluxDB       InfluxDBConfig       `json:"influxdb"`
	Auth           AuthConfig           `json:"auth"`
}

type ChargingEngineConfig struct {
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
	APIKey  string        `json:"api_key"`
}

type PaymentConfig struct {
	Provider            string `json:"provider"`
	StripeAPIKey        string `json:"stripe_api_key"`
	StripeWebhookSecret string `json:"stripe_webhook_secret"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	MetricsPort  string        `json:"metrics_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type IMSIConfig struct {
	Prefix  string `json:"prefix"`
	MinIMSI uint64 `json:"min_imsi"`
	MaxIMSI uint64 `json:"max_imsi"`
}

type InfluxDBConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
}

type ES2Config struct {
	BaseURL                  string `json:"base_url"`
	APIKey                   string `json:"api_key"`
	InsecureSkipVerify       bool   `json:"insecure_skip_verify"`
	FunctionalityRequesterID string `json:"functionality_requester_id"`
}

type AuthConfig struct {
	JWTSecret string `json:"jwt_secret"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("API_PORT", "8000"),
			MetricsPort:  getEnv("METRICS_PORT", "9090"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Database: getEnv("DB_NAME", "telecom_platform"),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		IMSI: IMSIConfig{
			Prefix:  "20893", // France MCC+MNC
			MinIMSI: 1,
			MaxIMSI: 999999999,
		},
		ES2: ES2Config{
			BaseURL:                  getEnv("ES2_BASE_URL", "https://smdp.example.com"),
			APIKey:                   getEnv("ES2_API_KEY", ""),
			InsecureSkipVerify:       true,
			FunctionalityRequesterID: "telecom-platform",
		},
		Payment: PaymentConfig{
			Provider:            getEnv("PAYMENT_PROVIDER", "stripe"),
			StripeAPIKey:        getEnv("STRIPE_API_KEY", ""),
			StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		ChargingEngine: ChargingEngineConfig{
			BaseURL: getEnv("CHARGING_ENGINE_URL", "http://localhost:3001"),
			Timeout: 10 * time.Second,
		},
		InfluxDB: InfluxDBConfig{
			URL:    getEnv("INFLUXDB_URL", "http://localhost:8086"),
			Token:  getEnv("INFLUXDB_TOKEN", ""),
			Org:    getEnv("INFLUXDB_ORG", "telecom"),
			Bucket: getEnv("INFLUXDB_BUCKET", "telecom"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
	}
}
