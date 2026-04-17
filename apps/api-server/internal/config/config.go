package config

import (
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
}

type ChargingEngineConfig struct {
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
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

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8000",
			MetricsPort:  "9090",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			Database: "telecom_platform",
			Username: "postgres",
			Password: "password",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		IMSI: IMSIConfig{
			Prefix:  "20893", // France MCC+MNC
			MinIMSI: 1,
			MaxIMSI: 999999999,
		},
		ES2: ES2Config{
			BaseURL:                  "https://smdp.example.com",
			APIKey:                   "",
			InsecureSkipVerify:       true,
			FunctionalityRequesterID: "telecom-platform",
		},
		Payment: PaymentConfig{
			Provider:            "stripe",
			StripeAPIKey:        "",
			StripeWebhookSecret: "",
		},
		ChargingEngine: ChargingEngineConfig{
			BaseURL: "http://localhost:3001",
			Timeout: 10 * time.Second,
		},
		InfluxDB: InfluxDBConfig{
			URL:    "http://localhost:8086",
			Token:  "",
			Org:    "telecom",
			Bucket: "telecom",
		},
	}
}
