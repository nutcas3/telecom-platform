package types

type CLIConfig struct {
	APIEndpoint string
	APIToken    string
	Profile     string
	Verbose     bool
	NoColor     bool
	Theme       string
}
