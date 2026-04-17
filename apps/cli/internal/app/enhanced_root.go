package app

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/api"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/commands"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/config"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

type EnhancedCLI struct {
	config *types.CLIConfig
}

// NewEnhancedCLI builds the CLI with defaults and attempts to load the persisted
// configuration from disk. The returned CLI always has a valid (defaulted)
// config, even if loading failed.
func NewEnhancedCLI() *EnhancedCLI {
	cfg := &types.CLIConfig{
		APIEndpoint: "http://localhost:8000",
		APIToken:    "",
		Profile:     "default",
		Verbose:     false,
		NoColor:     false,
		Theme:       "default",
	}

	// Attempt to hydrate from the viper-backed configuration.
	if persisted := loadPersistedConfig(); persisted != nil {
		if persisted.GetAPIEndpoint() != "" {
			cfg.APIEndpoint = persisted.GetAPIEndpoint()
		}
		if persisted.GetAPIToken() != "" {
			cfg.APIToken = persisted.GetAPIToken()
		}
		if persisted.Profile != "" {
			cfg.Profile = persisted.Profile
		}
		if persisted.UI.Theme != "" {
			cfg.Theme = persisted.UI.Theme
		}
		if persisted.NoColors() {
			cfg.NoColor = true
		}
	}

	return &EnhancedCLI{config: cfg}
}

// loadPersistedConfig is a best-effort load of the CLI configuration; errors
// are swallowed so the CLI still works with sane defaults.
func loadPersistedConfig() *config.Config {
	c := config.NewConfig()
	if err := c.Load(); err != nil {
		return c // still return defaults so callers can use GetAPIEndpoint()
	}
	return c
}

func (cli *EnhancedCLI) Run(args []string) error {
	// Parse any global flags first so subcommands see the updated config.
	_ = cli.parseConfig(args)

	if len(args) < 2 {
		cli.showHelp()
		return nil
	}

	command := args[1]
	commandArgs := args[2:]

	switch command {
	case "dashboard":
		return cli.runDashboard()
	case "subscribers":
		return commands.HandleSubscribersEnhanced(commandArgs, cli.config)
	case "services":
		return commands.HandleServicesEnhanced(commandArgs, cli.config)
	case "billing":
		return commands.HandleBillingEnhanced(commandArgs, cli.config)
	case "monitoring":
		return commands.HandleMonitoringEnhanced(commandArgs, cli.config)
	case "config":
		return commands.HandleConfigEnhanced(commandArgs, cli.config)
	case "deploy":
		return commands.HandleDeployEnhanced(commandArgs, cli.config)
	case "plugins":
		return commands.HandlePluginsEnhanced(commandArgs, cli.config)
	case "automation":
		return commands.HandleAutomationEnhanced(commandArgs, cli.config)
	case "help", "--help", "-h":
		cli.showHelp()
		return nil
	case "version", "--version", "-v":
		fmt.Println("telecom-cli v1.0.0")
		return nil
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		cli.showHelp()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (cli *EnhancedCLI) runDashboard() error {
	p := tea.NewProgram(NewDashboardWithConfig(cli.config))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running dashboard: %w", err)
	}
	return nil
}

// showHelp renders a lipgloss-styled help screen and an API connectivity banner
// so the user immediately sees whether the CLI can reach the backend.
func (cli *EnhancedCLI) showHelp() {
	useColor := !cli.config.NoColor

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cmdStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	render := func(s lipgloss.Style, text string) string {
		if !useColor {
			return text
		}
		return s.Render(text)
	}

	fmt.Println(render(headerStyle, "Telecom Platform CLI"))
	fmt.Println(render(mutedStyle, "Usage: telecom-cli <command> [options]"))
	fmt.Println()

	// API connectivity banner
	client := api.NewClient(cli.config)
	if client.IsConnected() {
		fmt.Println(render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
			"✓ Connected to API: "+cli.config.APIEndpoint,
		))
	} else {
		fmt.Println(render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true),
			"⚠ API unreachable at "+cli.config.APIEndpoint+" - commands will show sample data",
		))
	}
	fmt.Println()

	fmt.Println(render(subHeaderStyle, "Available commands:"))
	rows := [][2]string{
		{"dashboard", "Interactive dashboard for platform overview"},
		{"subscribers", "Subscriber management (list, show, create, delete, ...)"},
		{"services", "Service management (list, status, restart, logs)"},
		{"billing", "Billing and invoice management"},
		{"monitoring", "Monitoring, alerts, and health checks"},
		{"config", "Configuration management (show, get, set, validate)"},
		{"deploy", "Deployment management"},
		{"plugins", "Plugin management"},
		{"automation", "Automation and scripting"},
	}
	for _, r := range rows {
		fmt.Printf("  %-14s %s\n", render(cmdStyle, r[0]), render(mutedStyle, r[1]))
	}

	fmt.Println()
	fmt.Println(render(subHeaderStyle, "Global options:"))
	opts := [][2]string{
		{"--endpoint <url>", "API endpoint (default: http://localhost:8000)"},
		{"--token <token>", "API authentication token"},
		{"--profile <name>", "Configuration profile"},
		{"--verbose", "Enable verbose output"},
		{"--no-color", "Disable color output"},
	}
	for _, r := range opts {
		fmt.Printf("  %-22s %s\n", render(cmdStyle, r[0]), render(mutedStyle, r[1]))
	}

	fmt.Println()
	fmt.Println(render(mutedStyle, "Use 'telecom-cli <command>' with no args to see its subcommand help."))
}

func (cli *EnhancedCLI) parseConfig(args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--endpoint":
			if i+1 < len(args) {
				cli.config.APIEndpoint = args[i+1]
				i++
			}
		case "--token":
			if i+1 < len(args) {
				cli.config.APIToken = args[i+1]
				i++
			}
		case "--profile":
			if i+1 < len(args) {
				cli.config.Profile = args[i+1]
				i++
			}
		case "--verbose":
			cli.config.Verbose = true
		case "--no-color":
			cli.config.NoColor = true
		}
	}
	return nil
}

// Config returns a read-only view of the CLI configuration; primarily used by
// tests and by subcommands that need to bypass the command parser.
func (cli *EnhancedCLI) Config() *types.CLIConfig {
	return cli.config
}

func Main() {
	cli := NewEnhancedCLI()

	if err := cli.parseConfig(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing options: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
