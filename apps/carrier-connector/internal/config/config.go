package config

// ES2Config configures the upstream SM-DP+ ES2+ connection.
type ES2Config struct {
	BaseURL                  string `json:"base_url"`
	APIKey                   string `json:"api_key"`
	InsecureSkipVerify       bool   `json:"insecure_skip_verify"`
	FunctionalityRequesterID string `json:"functionality_requester_id"`
}

// DatabaseConfig configures the Postgres connection for the profile repository.
type DatabaseConfig struct {
	DSN string `json:"dsn"` // e.g. "host=localhost user=telecom password=... dbname=carrier port=5432 sslmode=disable"
}

// LoadES2Config returns default ES2+ configuration (override via env in main).
func LoadES2Config() *ES2Config {
	return &ES2Config{
		BaseURL:                  "https://smdp.example.com",
		APIKey:                   "",
		InsecureSkipVerify:       true,
		FunctionalityRequesterID: "telecom-platform",
	}
}
