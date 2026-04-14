package config

type ES2Config struct {
	BaseURL              string `json:"base_url"`
	APIKey               string `json:"api_key"`
	InsecureSkipVerify   bool   `json:"insecure_skip_verify"`
	FunctionalityRequesterID string `json:"functionality_requester_id"`
}

func LoadES2Config() *ES2Config {
	return &ES2Config{
		BaseURL:                "https://smdp.example.com",
		APIKey:                 "",
		InsecureSkipVerify:     true,
		FunctionalityRequesterID: "telecom-platform",
	}
}
