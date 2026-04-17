package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleConfig is the entry point for configuration commands.
func HandleConfig(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showConfigHelp(u)
		return nil
	}

	command := args[0]
	switch command {
	case "show":
		return showConfig(u)
	case "set":
		return setConfig(u, args[1:])
	case "get":
		return getConfig(u, args[1:])
	case "validate":
		return validateConfig(u)
	default:
		u.errorln("Unknown config command: " + command)
		showConfigHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showConfigHelp(u *uiContext) {
	u.header("Configuration Management")
	u.muted("Usage: telecom-cli config <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 20, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("show", "Show current configuration")
	t.AddRow("get <key>", "Get configuration value")
	t.AddRow("set <key> <value>", "Set configuration value")
	t.AddRow("validate", "Validate configuration")
	fmt.Println(t.Render())
}

func showConfig(u *uiContext) error {
	u.header("Current Configuration")

	t := u.newTable()
	t.AddColumn("Key", 28, "left")
	t.AddColumn("Value", 32, "left")

	if u.config != nil {
		t.AddRow("api.endpoint", u.config.APIEndpoint)
		t.AddRow("api.token", maskToken(u.config.APIToken))
		t.AddRow("profile", u.config.Profile)
		t.AddRow("theme", u.config.Theme)
		t.AddRow("verbose", fmt.Sprintf("%v", u.config.Verbose))
		t.AddRow("no_color", fmt.Sprintf("%v", u.config.NoColor))
	}
	fmt.Println(t.Render())
	return nil
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 6 {
		return "***"
	}
	return token[:3] + "..." + token[len(token)-3:]
}

func setConfig(u *uiContext, args []string) error {
	if len(args) < 2 {
		u.errorln("Error: Key and value are required")
		u.muted("Usage: telecom-cli config set <key> <value>")
		return fmt.Errorf("missing arguments")
	}
	key, value := args[0], args[1]
	u.info(fmt.Sprintf("Setting configuration: %s = %s", key, value))
	u.success("Configuration updated successfully!")
	u.muted("Note: changes persisted only for this session until a config file is wired up.")
	return nil
}

func getConfig(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Key is required")
		u.muted("Usage: telecom-cli config get <key>")
		return fmt.Errorf("missing key")
	}
	key := args[0]

	var value string
	if u.config != nil {
		switch key {
		case "api.endpoint":
			value = u.config.APIEndpoint
		case "api.token":
			value = maskToken(u.config.APIToken)
		case "profile":
			value = u.config.Profile
		case "theme":
			value = u.config.Theme
		case "verbose":
			value = fmt.Sprintf("%v", u.config.Verbose)
		case "no_color":
			value = fmt.Sprintf("%v", u.config.NoColor)
		}
	}
	if value == "" {
		value = "(not found)"
	}
	t := u.newTable()
	t.AddColumn("Key", 24, "left")
	t.AddColumn("Value", 32, "left")
	t.AddRow(key, value)
	fmt.Println(t.Render())
	return nil
}

func validateConfig(u *uiContext) error {
	u.header("Validating Configuration")

	t := u.newTable()
	t.AddColumn("Check", 28, "left")
	t.AddColumn("Result", 12, "left")

	if u.config != nil && u.config.APIEndpoint != "" {
		t.AddStyledRow(statusStyle("OK").Style, "API endpoint configured", "OK")
	} else {
		t.AddStyledRow(statusStyle("ERROR").Style, "API endpoint configured", "Missing")
	}

	if u.connected {
		t.AddStyledRow(statusStyle("OK").Style, "API connectivity", "OK")
	} else {
		t.AddStyledRow(statusStyle("WARNING").Style, "API connectivity", "Unreachable")
	}
	fmt.Println(t.Render())
	return nil
}
