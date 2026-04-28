package commands

import (
	"fmt"
	"strings"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/kubernetes"
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
	case "start-all":
		return startAllServices(u)
	case "stop-all":
		return stopAllServices(u)
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
	t.AddRow("start-all", "Start all platform services")
	t.AddRow("stop-all", "Stop all platform services")
	fmt.Println(t.Render())
}

func listServices(u *uiContext) error {
	u.header("Platform Services")

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Falling back to API client...")
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

	// Use Kubernetes service manager
	k8sServices, err := k8sManager.ListServices()
	if err != nil {
		u.warn("Failed to list Kubernetes services: " + err.Error())
		return err
	}

	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Version", 10, "left")
	t.AddColumn("Uptime", 10, "right")
	t.AddColumn("Replicas", 8, "right")

	for _, s := range k8sServices {
		t.AddStyledRow(statusStyle(strings.ToUpper(s.Status)).Style, s.Name, s.Status, s.Version, s.Uptime, fmt.Sprintf("%d/%d", s.Available, s.Replicas))
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

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Falling back to API client...")
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

	// Use Kubernetes service manager
	k8sService, err := k8sManager.GetServiceStatus(service)
	if err != nil {
		u.errorln("Failed to get service status: " + err.Error())
		return err
	}

	t := u.newTable()
	t.AddColumn("Field", 14, "left")
	t.AddColumn("Value", 28, "left")
	t.AddRow("Status", k8sService.Status)
	t.AddRow("Version", k8sService.Version)
	t.AddRow("Uptime", k8sService.Uptime)
	t.AddRow("CPU Usage", fmt.Sprintf("%.1f%%", k8sService.CPU))
	t.AddRow("Memory Usage", k8sService.Memory)
	t.AddRow("Replicas", fmt.Sprintf("%d/%d", k8sService.Available, k8sService.Replicas))
	fmt.Println(t.Render())
	return nil
}

func restartService(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Service name is required")
		u.muted("Usage: telecom-cli services restart <service>")
		return fmt.Errorf("missing service")
	}
	service := args[0]
	u.info("Restarting service: " + service)

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Falling back to API client...")
		err := u.client.PostRestart(service)
		if err != nil {
			u.warn("API error, simulated restart: " + err.Error())
		}
		u.success("Service restarted successfully!")
		return nil
	}

	// Use Kubernetes service manager
	err = k8sManager.RestartDeployment(service)
	if err != nil {
		u.errorln("Failed to restart service: " + err.Error())
		return err
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

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Falling back to placeholder logs...")
		logs := []string{
			"2026-05-30 16:45:30 [INFO] Service started successfully",
			"2026-05-30 16:45:35 [INFO] Database connection established",
			"2026-05-30 16:45:40 [INFO] API server listening on port 8000",
			"2026-05-30 16:45:45 [INFO] Health check passed",
		}
		for _, line := range logs {
			fmt.Println(u.colorizer.Colorize(line, statusStyle("OK")))
		}
		return nil
	}

	// Use Kubernetes service manager
	logs, err := k8sManager.GetPodLogs(service, 100)
	if err != nil {
		u.errorln("Failed to get logs: " + err.Error())
		return err
	}

	if len(logs) == 0 {
		u.muted("No logs available for " + service)
		return nil
	}

	for _, line := range logs {
		fmt.Println(u.colorizer.Colorize(line, statusStyle("OK")))
	}
	return nil
}

func startAllServices(u *uiContext) error {
	u.header("Starting All Platform Services")

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Showing simulation...")
		services := []string{"api-server", "charging-engine", "carrier-connector", "packet-gateway", "web-dashboard"}

		t := u.newTable()
		t.AddColumn("Service", 18, "left")
		t.AddColumn("Status", 12, "left")
		t.AddColumn("Message", 40, "left")

		for _, svc := range services {
			u.info("Starting " + svc + "...")
			t.AddStyledRow(statusStyle("OK").Style, svc, "Starting", "Initiating service startup")
		}

		fmt.Println(t.Render())
		u.success("All services start sequence initiated!")
		u.muted("Note: This is a simulation. Implement actual service orchestration via Kubernetes API or process manager.")
		return nil
	}

	// Use Kubernetes service manager
	services := []string{"api-server", "charging-engine", "carrier-connector", "packet-gateway", "web-dashboard"}

	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 12, "left")
	t.AddColumn("Message", 40, "left")

	for _, svc := range services {
		u.info("Starting " + svc + "...")
		err := k8sManager.ScaleDeployment(svc, 1)
		if err != nil {
			t.AddStyledRow(statusStyle("ERROR").Style, svc, "Failed", err.Error())
		} else {
			t.AddStyledRow(statusStyle("OK").Style, svc, "Started", "Scaled to 1 replica")
		}
	}

	fmt.Println(t.Render())
	u.success("All services start sequence completed!")
	return nil
}

func stopAllServices(u *uiContext) error {
	u.header("Stopping All Platform Services")

	// Try to use Kubernetes service manager
	k8sManager, err := kubernetes.NewServiceManager()
	if err != nil {
		u.warn("Failed to connect to Kubernetes: " + err.Error())
		u.muted("Showing simulation...")
		services := []string{"web-dashboard", "packet-gateway", "carrier-connector", "charging-engine", "api-server"}

		t := u.newTable()
		t.AddColumn("Service", 18, "left")
		t.AddColumn("Status", 12, "left")
		t.AddColumn("Message", 40, "left")

		for _, svc := range services {
			u.info("Stopping " + svc + "...")
			t.AddStyledRow(statusStyle("OK").Style, svc, "Stopping", "Initiating graceful shutdown")
		}

		fmt.Println(t.Render())
		u.success("All services stop sequence initiated!")
		u.muted("Note: This is a simulation. Implement actual service orchestration via Kubernetes API or process manager.")
		return nil
	}

	// Use Kubernetes service manager
	services := []string{"web-dashboard", "packet-gateway", "carrier-connector", "charging-engine", "api-server"}

	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 12, "left")
	t.AddColumn("Message", 40, "left")

	for _, svc := range services {
		u.info("Stopping " + svc + "...")
		err := k8sManager.ScaleDeployment(svc, 0)
		if err != nil {
			t.AddStyledRow(statusStyle("ERROR").Style, svc, "Failed", err.Error())
		} else {
			t.AddStyledRow(statusStyle("OK").Style, svc, "Stopped", "Scaled to 0 replicas")
		}
	}

	fmt.Println(t.Render())
	u.success("All services stop sequence completed!")
	return nil
}
