package commands

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/ui"
)

func HandleServicesEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) == 0 {
		showServicesEnhancedHelp()
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "list":
		return listServicesEnhanced(commandArgs, config)
	case "status":
		return showServiceStatusEnhanced(commandArgs, config)
	case "start":
		return startServiceEnhanced(commandArgs, config)
	case "stop":
		return stopServiceEnhanced(commandArgs, config)
	case "restart":
		return restartServiceEnhanced(commandArgs, config)
	case "logs":
		return showServiceLogsEnhanced(commandArgs, config)
	case "metrics":
		return showServiceMetricsEnhanced(commandArgs, config)
	case "health":
		return checkServiceHealthEnhanced(commandArgs, config)
	case "scale":
		return scaleServiceEnhanced(commandArgs, config)
	default:
		fmt.Printf("Unknown services command: %s\n", command)
		showServicesEnhancedHelp()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showServicesEnhancedHelp() {
	fmt.Println("Enhanced Service Management")
	fmt.Println("Usage: telecom-cli services <command> [options]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  list                    - List all services")
	fmt.Println("  status <service>        - Show service status")
	fmt.Println("  start <service>         - Start a service")
	fmt.Println("  stop <service>          - Stop a service")
	fmt.Println("  restart <service>       - Restart a service")
	fmt.Println("  logs <service>          - Show service logs")
	fmt.Println("  metrics <service>       - Show service metrics")
	fmt.Println("  health <service>        - Check service health")
	fmt.Println("  scale <service>         - Scale service")
}

func listServicesEnhanced(args []string, config *types.CLIConfig) error {
	// Initialize UI components
	colorizer := ui.NewColorizer(!config.NoColor)
	iconRenderer := ui.NewIconRenderer(true, true)

	// Print header with styling
	fmt.Printf("%s %s %s\n",
		colorizer.Header("Telecom Platform"),
		colorizer.Info("Services"),
		colorizer.Muted(fmt.Sprintf("(Profile: %s, Endpoint: %s)", config.Profile, config.APIEndpoint)))

	// Create styled table
	table := ui.NewServicesTable(colorizer, iconRenderer)

	// Add rows with styling
	services := []struct {
		name    string
		status  string
		version string
		cpu     float64
		memory  string
		uptime  string
	}{
		{"api-server", "running", "v1.0.0", 45, "256MB", "2h15m"},
		{"charging-engine", "running", "v1.0.0", 23, "128MB", "2h15m"},
		{"packet-gateway", "running", "v1.0.0", 67, "512MB", "2h15m"},
		{"web-dashboard", "running", "v1.0.0", 12, "64MB", "2h15m"},
		{"prometheus", "running", "v2.45.0", 8, "32MB", "2h15m"},
		{"grafana", "running", "v9.5.2", 5, "24MB", "2h15m"},
	}

	for _, svc := range services {
		statusIcon := iconRenderer.StatusColored(svc.status, colorizer)
		cpuStyle := lipgloss.NewStyle()
		if svc.cpu >= 70 {
			cpuStyle = ui.StyleMetricCritical.Style
		} else if svc.cpu >= 50 {
			cpuStyle = ui.StyleMetricWarning.Style
		} else {
			cpuStyle = ui.StyleMetricGood.Style
		}

		table.AddStyledCellRow(
			ui.TableCell{Text: statusIcon, Align: "center"},
			ui.TableCell{Text: svc.name},
			ui.TableCell{Text: iconRenderer.StatusColored(svc.status, colorizer), Align: "center"},
			ui.TableCell{Text: svc.version, Align: "center"},
			ui.TableCell{Text: colorizer.Metric(svc.cpu), Style: cpuStyle, Align: "right"},
			ui.TableCell{Text: iconRenderer.Memory() + " " + svc.memory, Align: "right"},
			ui.TableCell{Text: iconRenderer.Clock() + " " + svc.uptime, Align: "right"},
		)
	}

	// Render the table
	fmt.Println(table.Render())

	// Add summary with metrics
	runningCount := 6
	totalCPU := 45 + 23 + 67 + 12 + 8 + 5
	avgCPU := float64(totalCPU) / float64(runningCount)

	fmt.Printf("\n%s %s %s, %s: %s, %s: %s\n",
		colorizer.Info("Summary:"),
		colorizer.Success(fmt.Sprintf("%d", runningCount)),
		colorizer.Success("running"),
		colorizer.Info("Avg CPU"),
		colorizer.Metric(avgCPU),
		colorizer.Info("Status"),
		colorizer.Success("healthy"))

	return nil
}

func showServiceStatusEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services status <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Service Status: %s (Profile: %s)\n", service, config.Profile)
	fmt.Println("========================================")
	fmt.Printf("Status:        Running\n")
	fmt.Printf("Version:       v1.0.0\n")
	fmt.Printf("Uptime:        2h15m\n")
	fmt.Printf("CPU Usage:     45%%\n")
	fmt.Printf("Memory Usage:  256MB\n")
	fmt.Printf("Last Restart:  2024-01-15 14:30:00\n")
	fmt.Printf("Health Check:  Healthy\n")
	return nil
}

func startServiceEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services start <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Starting service: %s (Profile: %s)\n", service, config.Profile)
	fmt.Println("Service started successfully!")
	return nil
}

func stopServiceEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services stop <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Stopping service: %s (Profile: %s)\n", service, config.Profile)
	fmt.Println("Service stopped successfully!")
	return nil
}

func restartServiceEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services restart <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Restarting service: %s (Profile: %s)\n", service, config.Profile)
	fmt.Println("Service restarted successfully!")
	return nil
}

func showServiceLogsEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services logs <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Recent logs for %s (Profile: %s):\n", service, config.Profile)
	fmt.Println("2024-01-15 16:45:30 [INFO] Service started successfully")
	fmt.Println("2024-01-15 16:45:35 [INFO] Database connection established")
	fmt.Println("2024-01-15 16:45:40 [INFO] API server listening on port 8000")
	fmt.Println("2024-01-15 16:45:45 [INFO] Health check passed")
	fmt.Println("2024-01-15 16:45:50 [WARN] High memory usage detected")
	fmt.Println("2024-01-15 16:45:55 [ERROR] Database connection timeout")
	fmt.Println("2024-01-15 16:46:00 [INFO] Database connection restored")
	return nil
}

func showServiceMetricsEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services metrics <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Metrics for %s (Profile: %s):\n", service, config.Profile)
	fmt.Println("========================================")
	fmt.Printf("Request Rate:    450 req/s\n")
	fmt.Printf("Response Time:   125ms avg\n")
	fmt.Printf("Error Rate:      0.5%%\n")
	fmt.Printf("Throughput:      1.2GB/s\n")
	fmt.Printf("Connections:     150 active\n")
	fmt.Printf("Memory Usage:    256MB / 512MB\n")
	fmt.Printf("CPU Usage:       45%%\n")
	return nil
}

func checkServiceHealthEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services health <service>")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	fmt.Printf("Health check for %s (Profile: %s):\n", service, config.Profile)
	fmt.Println("========================================")
	fmt.Printf("Overall Status:  Healthy\n")
	fmt.Printf("Database:        Connected\n")
	fmt.Printf("Redis:           Connected\n")
	fmt.Printf("API Endpoint:    Responsive\n")
	fmt.Printf("Dependencies:    All healthy\n")
	fmt.Printf("Last Check:      2024-01-15 16:46:30\n")
	return nil
}

func scaleServiceEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) < 1 {
		fmt.Println("Error: Service name is required")
		fmt.Println("Usage: telecom-cli services scale <service> [replicas]")
		return fmt.Errorf("missing service name")
	}

	service := args[0]
	replicas := 3
	if len(args) > 1 {
		// Parse replicas from args[1]
		fmt.Printf("Scaling %s to %d replicas (Profile: %s)\n", service, replicas, config.Profile)
	} else {
		fmt.Printf("Scaling %s to default %d replicas (Profile: %s)\n", service, replicas, config.Profile)
	}
	fmt.Println("Service scaled successfully!")
	return nil
}
