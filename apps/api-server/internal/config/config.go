package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	IMSI     IMSIConfig      `json:"imsi"`
	ES2      ES2Config       `json:"es2"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
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
	Prefix string `json:"prefix"`
	MinIMSI uint64 `json:"min_imsi"`
	MaxIMSI uint64 `json:"max_imsi"`
}

type ES2Config struct {
	BaseURL              string `json:"base_url"`
	APIKey               string `json:"api_key"`
	InsecureSkipVerify   bool   `json:"insecure_skip_verify"`
	FunctionalityRequesterID string `json:"functionality_requester_id"`
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8000",
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
			Prefix: "20893", // France MCC+MNC
			MinIMSI: 1,
			MaxIMSI: 999999999,
		},
		ES2: ES2Config{
			BaseURL:                "https://smdp.example.com",
			APIKey:                 "",
			InsecureSkipVerify:     true,
			FunctionalityRequesterID: "telecom-platform",
		},
	}
}
