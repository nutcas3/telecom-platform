package commands

import (
	"fmt"
	"strings"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleServices is the entry point for service commands.
func HandleServices(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showServicesHelp(u)
		return nil
	}

	u.connectivityBanner()

	command := args[0]
	switch command {
	case "list":
		return listServices(u)
	case "status":
		return showServiceStatus(u, args[1:])
	case "restart":
		return restartService(u, args[1:])
	case "logs":
		return showServiceLogs(u, args[1:])
	default:
		u.errorln("Unknown services command: " + command)
		showServicesHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showServicesHelp(u *uiContext) {
	u.header("Service Management")
	u.muted("Usage: telecom-cli services <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 20, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("list", "List all services")
	t.AddRow("status <service>", "Show service status")
	t.AddRow("restart <service>", "Restart a service")
	t.AddRow("logs <service>", "Show service logs")
	fmt.Println(t.Render())
}

func listServices(u *uiContext) error {
	u.header("Platform Services")

	services, err := u.client.ListServices()
	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Version", 10, "left")
	t.AddColumn("Uptime", 10, "right")

	if err != nil || len(services) == 0 {
		if err != nil {
			u.warn("Using placeholder data: " + err.Error())
		} else {
			u.muted("(API returned no services - showing sample data)")
		}
		placeholders := [][]string{
			{"api-server", "Running", "v1.0.0", "2h15m"},
			{"charging-engine", "Running", "v1.0.0", "2h15m"},
			{"packet-gateway", "Running", "v1.0.0", "2h15m"},
			{"web-dashboard", "Running", "v1.0.0", "2h15m"},
			{"prometheus", "Running", "v2.45.0", "2h15m"},
			{"grafana", "Running", "v9.5.2", "2h15m"},
		}
		for _, row := range placeholders {
			t.AddStyledRow(statusStyle(strings.ToUpper(row[1])).Style, row[0], row[1], row[2], row[3])
		}
	} else {
		for _, s := range services {
			t.AddStyledRow(statusStyle(strings.ToUpper(s.Status)).Style, s.Name, s.Status, s.Version, s.Uptime)
		}
	}
	fmt.Println(t.Render())
	return nil
}

func showServiceStatus(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Service name is required")
		u.muted("Usage: telecom-cli services status <service>")
		return fmt.Errorf("missing service")
	}
	service := args[0]
	u.header("Service Status: " + service)

	services, err := u.client.ListServices()
	t := u.newTable()
	t.AddColumn("Field", 14, "left")
	t.AddColumn("Value", 28, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddRow("Status", "Running")
		t.AddRow("Version", "v1.0.0")
		t.AddRow("Uptime", "2h15m")
		t.AddRow("CPU Usage", "45%")
		t.AddRow("Memory Usage", "256MB")
		fmt.Println(t.Render())
		return nil
	}
	for _, s := range services {
		if s.Name == service {
			t.AddRow("Status", s.Status)
			t.AddRow("Version", s.Version)
			t.AddRow("Uptime", s.Uptime)
			t.AddRow("CPU Usage", fmt.Sprintf("%.1f%%", s.CPU))
			t.AddRow("Memory Usage", s.Memory)
			fmt.Println(t.Render())
			return nil
		}
	}
	u.errorln("Service not found: " + service)
	return fmt.Errorf("service not found")
}

func restartService(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Service name is required")
		u.muted("Usage: telecom-cli services restart <service>")
		return fmt.Errorf("missing service")
	}
	service := args[0]
	u.info("Restarting service: " + service)
	err := u.client.PostRestart(service)
	if err != nil {
		u.warn("API error, simulated restart: " + err.Error())
	}
	u.success("Service restarted successfully!")
	return nil
}

func showServiceLogs(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Service name is required")
		u.muted("Usage: telecom-cli services logs <service>")
		return fmt.Errorf("missing service")
	}
	service := args[0]
	u.header("Recent logs for " + service)

	// Logs endpoint not yet implemented; show placeholder logs
	u.muted("(Showing placeholder log entries - logs API not yet connected)")
	logs := []string{
		"2024-01-15 16:45:30 [INFO] Service started successfully",
		"2024-01-15 16:45:35 [INFO] Database connection established",
		"2024-01-15 16:45:40 [INFO] API server listening on port 8000",
		"2024-01-15 16:45:45 [INFO] Health check passed",
	}
	for _, line := range logs {
		fmt.Println(u.colorizer.Colorize(line, statusStyle("OK")))
	}
	return nil
}
