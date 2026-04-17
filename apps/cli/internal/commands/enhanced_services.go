package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleServicesEnhanced delegates to the UI + API-connected services handler.
// Additional subcommands (start/stop/scale/health) use UI-styled placeholders
// until their API endpoints are available.
func HandleServicesEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) == 0 {
		return HandleServices(args, config)
	}

	command := args[0]
	switch command {
	case "list", "status", "restart", "logs":
		return HandleServices(args, config)
	case "start", "stop":
		return lifecycleServiceUI(command, args[1:], config)
	case "scale":
		return scaleServiceUI(args[1:], config)
	case "health":
		return healthServiceUI(args[1:], config)
	default:
		u := newUIContext(config)
		u.errorln("Unknown services command: " + command)
		return HandleServices(nil, config)
	}
}

func lifecycleServiceUI(cmd string, args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: Service name is required")
		u.muted("Usage: telecom-cli services " + cmd + " <service>")
		return fmt.Errorf("missing service")
	}
	service := args[0]
	u.info(cmd + " service: " + service)
	u.muted("Lifecycle API endpoint not yet wired; showing placeholder confirmation.")
	u.success("Service " + service + " " + cmd + " completed (simulated)")
	return nil
}

func scaleServiceUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 2 {
		u.errorln("Error: Service name and replica count are required")
		u.muted("Usage: telecom-cli services scale <service> <replicas>")
		return fmt.Errorf("missing arguments")
	}
	service, replicas := args[0], args[1]
	u.info("Scaling " + service + " to " + replicas + " replicas")
	u.muted("Scaling API endpoint not yet wired; showing placeholder confirmation.")
	u.success("Service " + service + " scaled (simulated)")
	return nil
}

func healthServiceUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}
	u.header("Service Health: " + target)

	health, err := u.client.GetHealth()
	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Uptime", 10, "right")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddStyledRow(statusStyle("HEALTHY").Style, "api-server", "Healthy", "99.9%")
		t.AddStyledRow(statusStyle("HEALTHY").Style, "charging-engine", "Healthy", "99.8%")
		fmt.Println(t.Render())
		return nil
	}
	for _, h := range health {
		if target != "all" && h.Service != target {
			continue
		}
		t.AddStyledRow(statusStyle(h.Status).Style, h.Service, h.Status, h.Uptime)
	}
	fmt.Println(t.Render())
	return nil
}
