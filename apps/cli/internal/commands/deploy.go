package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleDeploy is the entry point for deployment commands.
func HandleDeploy(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showDeployHelp(u)
		return nil
	}

	u.connectivityBanner()

	command := args[0]
	switch command {
	case "status":
		return showDeployStatus(u)
	case "start":
		return startDeployment(u, args[1:])
	case "rollback":
		return rollbackDeployment(u, args[1:])
	case "history":
		return showDeployHistory(u)
	default:
		u.errorln("Unknown deploy command: " + command)
		showDeployHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showDeployHelp(u *uiContext) {
	u.header("Deployment Management")
	u.muted("Usage: telecom-cli deploy <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 22, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("status", "Show deployment status")
	t.AddRow("start <environment>", "Start deployment")
	t.AddRow("rollback <version>", "Rollback to version")
	t.AddRow("history", "Show deployment history")
	fmt.Println(t.Render())
}

func showDeployStatus(u *uiContext) error {
	u.header("Deployment Status")
	u.muted("(Deployment API not yet connected - showing placeholder)")

	t := u.newTable()
	t.AddColumn("Environment", 14, "left")
	t.AddColumn("Status", 12, "left")
	t.AddColumn("Version", 10, "left")
	t.AddColumn("Last Deploy", 18, "left")
	t.AddStyledRow(statusStyle("HEALTHY").Style, "production", "Healthy", "v1.0.0", "2026-05-30 14:30")
	t.AddStyledRow(statusStyle("HEALTHY").Style, "staging", "Healthy", "v1.0.1", "2026-05-30 13:45")
	t.AddStyledRow(statusStyle("BUILDING").Style, "development", "Building", "v1.1.0", "2026-05-30 12:00")
	fmt.Println(t.Render())
	return nil
}

func startDeployment(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Environment is required")
		u.muted("Usage: telecom-cli deploy start <environment>")
		return fmt.Errorf("missing environment")
	}
	env := args[0]
	u.info("Starting deployment to " + env + "...")
	steps := []string{
		"Building application...",
		"Running tests...",
		"Creating deployment package...",
		"Deploying to Kubernetes...",
	}
	for _, s := range steps {
		u.muted(s)
	}
	u.success("Deployment to " + env + " completed successfully!")
	return nil
}

func rollbackDeployment(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Version is required")
		u.muted("Usage: telecom-cli deploy rollback <version>")
		return fmt.Errorf("missing version")
	}
	version := args[0]
	u.info("Rolling back to version " + version + "...")
	steps := []string{
		"Stopping current deployment...",
		"Deploying previous version...",
		"Running health checks...",
	}
	for _, s := range steps {
		u.muted(s)
	}
	u.success("Rollback to " + version + " completed successfully!")
	return nil
}

func showDeployHistory(u *uiContext) error {
	u.header("Deployment History")
	u.muted("(Deployment API not yet connected - showing placeholder)")

	t := u.newTable()
	t.AddColumn("Version", 10, "left")
	t.AddColumn("Environment", 14, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Date", 12, "left")
	t.AddColumn("Time", 8, "left")

	rows := [][]string{
		{"v1.0.0", "production", "Success", "2026-05-30", "14:30"},
		{"v1.0.1", "staging", "Success", "2026-05-30", "13:45"},
		{"v1.0.0", "staging", "Success", "2026-05-30", "12:30"},
		{"v0.9.9", "production", "Failed", "2024-01-14", "16:20"},
		{"v0.9.8", "production", "Success", "2024-01-14", "15:45"},
		{"v0.9.7", "production", "Success", "2024-01-14", "14:15"},
	}
	for _, r := range rows {
		style := statusStyle("SUCCESS")
		if r[2] == "Failed" {
			style = statusStyle("FAILED")
		}
		t.AddStyledRow(style.Style, r[0], r[1], r[2], r[3], r[4])
	}
	fmt.Println(t.Render())
	return nil
}
