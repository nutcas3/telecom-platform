package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ConfigFile  string
	Profile     string
	APIEndpoint string
	APIToken    string
	Verbose     bool
	NoColor     bool

	// Configuration sections
	API     APIConfig     `mapstructure:"api"`
	UI      UIConfig      `mapstructure:"ui"`
	Plugins PluginConfig  `mapstructure:"plugins"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type APIConfig struct {
	Endpoint string            `mapstructure:"endpoint"`
	Timeout  string            `mapstructure:"timeout"`
	Retries  int               `mapstructure:"retries"`
	Token    string            `mapstructure:"token"`
	Insecure bool              `mapstructure:"insecure"`
	Headers  map[string]string `mapstructure:"headers"`
}

type UIConfig struct {
	Theme       string `mapstructure:"theme"`
	Colors      bool   `mapstructure:"colors"`
	Pager       bool   `mapstructure:"pager"`
	TableStyle  string `mapstructure:"table-style"`
	ChartWidth  int    `mapstructure:"chart-width"`
	ChartHeight int    `mapstructure:"chart-height"`
}

type PluginConfig struct {
	Enabled  []string       `mapstructure:"enabled"`
	Disabled []string       `mapstructure:"disabled"`
	Config   map[string]any `mapstructure:"config"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max-size"`
	MaxBackups int    `mapstructure:"max-backups"`
	MaxAge     int    `mapstructure:"max-age"`
}

func NewConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	configDir := filepath.Join(home, ".telecom-cli")
	defaultConfigFile := filepath.Join(configDir, "config.yaml")

	return &Config{
		ConfigFile: defaultConfigFile,
		API: APIConfig{
			Endpoint: "http://localhost:8000",
			Timeout:  "30s",
			Retries:  3,
			Insecure: false,
			Headers:  make(map[string]string),
		},
		UI: UIConfig{
			Theme:       "default",
			Colors:      true,
			Pager:       true,
			TableStyle:  "default",
			ChartWidth:  80,
			ChartHeight: 20,
		},
		Plugins: PluginConfig{
			Enabled:  []string{},
			Disabled: []string{},
			Config:   make(map[string]any),
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
	}
}

func (c *Config) Load() error {
	// Set defaults
	viper.SetDefault("api.endpoint", "http://localhost:8000")
	viper.SetDefault("api.timeout", "30s")
	viper.SetDefault("api.retries", 3)
	viper.SetDefault("ui.theme", "default")
	viper.SetDefault("ui.colors", true)
	viper.SetDefault("ui.pager", true)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.output", "stdout")

	// Set config file path
	if c.ConfigFile != "" {
		viper.SetConfigFile(c.ConfigFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		configDir := filepath.Join(home, ".telecom-cli")
		configFile := filepath.Join(configDir, "config.yaml")
		viper.SetConfigFile(configFile)
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("TELECOM_CLI")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			if err := c.createDefaultConfig(); err != nil {
				return fmt.Errorf("could not create default config: %w", err)
			}
		} else {
			return fmt.Errorf("could not read config file: %w", err)
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("could not unmarshal config: %w", err)
	}

	// Override with command line flags
	if c.APIEndpoint != "" {
		c.API.Endpoint = c.APIEndpoint
	}
	if c.APIToken != "" {
		c.API.Token = c.APIToken
	}

	// Apply profile if specified
	if c.Profile != "" {
		if err := c.applyProfile(c.Profile); err != nil {
			return fmt.Errorf("could not apply profile '%s': %w", c.Profile, err)
		}
	}

	return nil
}

func (c *Config) createDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".telecom-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Create default config
	defaultConfig := map[string]any{
		"api": map[string]any{
			"endpoint": "http://localhost:8000",
			"timeout":  "30s",
			"retries":  3,
			"insecure": false,
		},
		"ui": map[string]any{
			"theme":       "default",
			"colors":      true,
			"pager":       true,
			"table-style": "default",
		},
		"logging": map[string]any{
			"level":  "info",
			"format": "text",
			"output": "stdout",
		},
		"plugins": map[string]any{
			"enabled":  []string{},
			"disabled": []string{},
		},
		"profiles": map[string]any{
			"development": map[string]any{
				"api": map[string]any{
					"endpoint": "http://dev-api:8000",
					"insecure": true,
				},
				"logging": map[string]any{
					"level": "debug",
				},
			},
			"production": map[string]any{
				"api": map[string]any{
					"endpoint": "https://api.telecom-platform.com",
					"timeout":  "60s",
				},
				"logging": map[string]any{
					"level": "warn",
				},
			},
		},
	}

	viper.Set("default", defaultConfig)
	return viper.WriteConfigAs(configFile)
}

func (c *Config) applyProfile(profileName string) error {
	profiles := viper.GetStringMap("profiles")
	profileData, exists := profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Apply profile configuration
	if profile, ok := profileData.(map[string]any); ok {
		// Apply API settings
		if api, ok := profile["api"].(map[string]any); ok {
			if endpoint, ok := api["endpoint"].(string); ok {
				c.API.Endpoint = endpoint
			}
			if timeout, ok := api["timeout"].(string); ok {
				c.API.Timeout = timeout
			}
			if retries, ok := api["retries"].(int); ok {
				c.API.Retries = retries
			}
			if insecure, ok := api["insecure"].(bool); ok {
				c.API.Insecure = insecure
			}
		}

		// Apply UI settings
		if ui, ok := profile["ui"].(map[string]any); ok {
			if theme, ok := ui["theme"].(string); ok {
				c.UI.Theme = theme
			}
			if colors, ok := ui["colors"].(bool); ok {
				c.UI.Colors = colors
			}
		}

		// Apply logging settings
		if logging, ok := profile["logging"].(map[string]any); ok {
			if level, ok := logging["level"].(string); ok {
				c.Logging.Level = level
			}
			if format, ok := logging["format"].(string); ok {
				c.Logging.Format = format
			}
		}
	}

	return nil
}

func (c *Config) Save() error {
	if c.ConfigFile == "" {
		return fmt.Errorf("no config file specified")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(c.ConfigFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	viper.Set("api", c.API)
	viper.Set("ui", c.UI)
	viper.Set("plugins", c.Plugins)
	viper.Set("logging", c.Logging)

	return viper.WriteConfigAs(c.ConfigFile)
}

func (c *Config) GetAPIEndpoint() string {
	if c.APIEndpoint != "" {
		return c.APIEndpoint
	}
	return c.API.Endpoint
}

func (c *Config) GetAPIToken() string {
	if c.APIToken != "" {
		return c.APIToken
	}
	return c.API.Token
}

func (c *Config) IsVerbose() bool {
	return c.Verbose
}

func (c *Config) NoColors() bool {
	return c.NoColor || !c.UI.Colors
}
