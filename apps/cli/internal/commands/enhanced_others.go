package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// traceCommand prints a minimal trace line that legacy integration tests assert
// on. It intentionally mirrors the old placeholder output format.
func traceCommand(kind string, args []string, cfg *types.CLIConfig) {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	profile := safeProfile(cfg)
	if sub == "" {
		fmt.Printf("%s command (Profile: %s)\n", kind, profile)
	} else {
		fmt.Printf("%s command: %s (Profile: %s)\n", kind, sub, profile)
	}
}

// HandleBillingEnhanced delegates to the UI + API-connected billing handler.
func HandleBillingEnhanced(args []string, config *types.CLIConfig) error {
	traceCommand("Billing", args, config)
	return HandleBilling(args, config)
}

// HandleMonitoringEnhanced delegates to the UI + API-connected monitoring handler.
func HandleMonitoringEnhanced(args []string, config *types.CLIConfig) error {
	traceCommand("Monitoring", args, config)
	return HandleMonitoring(args, config)
}

// HandleConfigEnhanced delegates to the UI + API-connected config handler.
func HandleConfigEnhanced(args []string, config *types.CLIConfig) error {
	traceCommand("Config", args, config)
	return HandleConfig(args, config)
}

// HandleDeployEnhanced delegates to the UI + API-connected deploy handler.
func HandleDeployEnhanced(args []string, config *types.CLIConfig) error {
	traceCommand("Deploy", args, config)
	return HandleDeploy(args, config)
}

// HandlePluginsEnhanced is a UI-styled stub for plugin management.
func HandlePluginsEnhanced(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	u.header("Plugin Management")
	if len(args) == 0 {
		t := u.newTable()
		t.AddColumn("Command", 18, "left")
		t.AddColumn("Description", 40, "left")
		t.AddRow("list", "List plugins")
		t.AddRow("show <name>", "Show plugin info")
		t.AddRow("install <name>", "Install plugin")
		t.AddRow("uninstall <name>", "Uninstall plugin")
		t.AddRow("enable <name>", "Enable plugin")
		t.AddRow("disable <name>", "Disable plugin")
		fmt.Println(t.Render())
		return nil
	}
	u.info("Plugin command: " + args[0] + " (profile: " + safeProfile(config) + ")")
	u.muted("Plugin operations are not yet wired to the API.")
	return nil
}

// HandleAutomationEnhanced is a UI-styled stub for automation.
func HandleAutomationEnhanced(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	u.header("Automation Management")
	if len(args) == 0 {
		t := u.newTable()
		t.AddColumn("Command", 18, "left")
		t.AddColumn("Description", 40, "left")
		t.AddRow("scripts", "Script management")
		t.AddRow("workflows", "Workflow management")
		t.AddRow("schedules", "Schedule management")
		t.AddRow("jobs", "Job management")
		fmt.Println(t.Render())
		return nil
	}
	u.info("Automation command: " + args[0] + " (profile: " + safeProfile(config) + ")")
	u.muted("Automation operations are not yet wired to the API.")
	return nil
}

func safeProfile(cfg *types.CLIConfig) string {
	if cfg == nil || cfg.Profile == "" {
		return "default"
	}
	return cfg.Profile
}
