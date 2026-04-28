package commands

import (
	"fmt"
	"strings"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleMonitoring is the entry point for monitoring commands.
func HandleMonitoring(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showMonitoringHelp(u)
		return nil
	}

	u.connectivityBanner()

	command := args[0]
	switch command {
	case "metrics":
		return showMetrics(u)
	case "alerts":
		return showAlerts(u)
	case "health":
		return showHealth(u)
	case "check-all":
		return checkAllServices(u)
	case "logs":
		return showLogs(u, args[1:])
	default:
		u.errorln("Unknown monitoring command: " + command)
		showMonitoringHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showMonitoringHelp(u *uiContext) {
	u.header("Monitoring Management")
	u.muted("Usage: telecom-cli monitoring <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 18, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("metrics", "Show system metrics")
	t.AddRow("alerts", "Show active alerts")
	t.AddRow("health", "Show system health")
	t.AddRow("check-all", "Check health of all services")
	t.AddRow("logs <service>", "Show service logs")
	fmt.Println(t.Render())
}

func showMetrics(u *uiContext) error {
	u.header("System Metrics")

	stats, err := u.client.GetSystemStats()
	t := u.newTable()
	t.AddColumn("Metric", 24, "left")
	t.AddColumn("Value", 16, "right")
	t.AddColumn("Status", 10, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		rows := [][]string{
			{"API Response Time", "125ms", "Good"},
			{"Database Connections", "15/100", "Good"},
			{"CPU Usage", "45%", "Good"},
			{"Memory Usage", "2.3GB/8GB", "Good"},
			{"Active Subscribers", "1234", "Good"},
			{"Request Rate", "450/s", "Good"},
		}
		for _, r := range rows {
			t.AddStyledRow(statusStyle("HEALTHY").Style, r[0], r[1], r[2])
		}
		fmt.Println(t.Render())
		return nil
	}
	t.AddRow("Active Sessions", fmt.Sprintf("%d", stats.ActiveSessions))
	t.AddRow("Total Accounts", fmt.Sprintf("%d", stats.TotalAccounts))
	t.AddRow("Blocked Users", fmt.Sprintf("%d", stats.BlockedUsers))
	t.AddRow("Low Balance Alerts", fmt.Sprintf("%d", stats.LowBalanceAlerts))
	t.AddRow("CPU Usage", fmt.Sprintf("%.1f%%", stats.CPUUsage))
	t.AddRow("Memory Usage", fmt.Sprintf("%.1f%%", stats.MemoryUsage))
	t.AddRow("Uptime", stats.Uptime)
	fmt.Println(t.Render())
	return nil
}

func showAlerts(u *uiContext) error {
	u.header("Active Alerts")

	alerts, err := u.client.ListAlerts()
	t := u.newTable()
	t.AddColumn("Severity", 10, "left")
	t.AddColumn("Service", 16, "left")
	t.AddColumn("Message", 40, "left")
	t.AddColumn("Time", 10, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		rows := [][]string{
			{"High", "API Server", "High memory usage detected", "14:30"},
			{"Medium", "Database", "Slow query detected", "14:25"},
			{"Low", "Gateway", "Packet loss increased", "14:20"},
		}
		for _, r := range rows {
			style := statusStyle(strings.ToUpper(r[0]))
			switch r[0] {
			case "High":
				style = statusStyle("ERROR")
			case "Medium":
				style = statusStyle("WARNING")
			}
			t.AddStyledRow(style.Style, r[0], r[1], r[2], r[3])
		}
		fmt.Println(t.Render())
		return nil
	}
	for _, a := range alerts {
		style := statusStyle(strings.ToUpper(a.Severity))
		t.AddStyledRow(style.Style, a.Severity, a.Service, a.Message, a.Time.Format("15:04"))
	}
	fmt.Println(t.Render())
	return nil
}

func showHealth(u *uiContext) error {
	u.header("System Health Status")

	health, err := u.client.GetHealth()
	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Uptime", 10, "right")
	t.AddColumn("Last Check", 12, "right")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		rows := [][]string{
			{"API Server", "Healthy", "99.9%", "14:45:30"},
			{"Charging Engine", "Healthy", "99.8%", "14:45:30"},
			{"Packet Gateway", "Healthy", "99.7%", "14:45:30"},
			{"Web Dashboard", "Healthy", "99.9%", "14:45:30"},
			{"Database", "Healthy", "99.5%", "14:45:30"},
			{"Redis", "Healthy", "99.9%", "14:45:30"},
		}
		for _, r := range rows {
			t.AddStyledRow(statusStyle(strings.ToUpper(r[1])).Style, r[0], r[1], r[2], r[3])
		}
		fmt.Println(t.Render())
		return nil
	}
	for _, h := range health {
		t.AddStyledRow(statusStyle(strings.ToUpper(h.Status)).Style,
			h.Service, h.Status, h.Uptime, h.LastCheck.Format("15:04:05"))
	}
	fmt.Println(t.Render())
	return nil
}

func checkAllServices(u *uiContext) error {
	u.header("All Services Health Check")

	healthResults, err := u.client.CheckAllServices()
	t := u.newTable()
	t.AddColumn("Service", 18, "left")
	t.AddColumn("Status", 12, "left")
	t.AddColumn("Endpoint", 25, "left")
	t.AddColumn("Message", 30, "left")

	if err != nil {
		u.errorln("Failed to check services: " + err.Error())
		return err
	}

	allHealthy := true
	for _, h := range healthResults {
		style := statusStyle(strings.ToUpper(h.Status))
		if h.Status != "healthy" {
			allHealthy = false
		}
		t.AddStyledRow(style.Style, h.Name, h.Status, h.Endpoint, h.Message)
	}
	fmt.Println(t.Render())

	if allHealthy {
		u.success("All services are healthy")
	} else {
		u.warn("Some services are unhealthy or unreachable")
	}
	return nil
}

func showLogs(u *uiContext, args []string) error {
	service := "all"
	if len(args) > 0 {
		service = args[0]
	}
	u.header("Recent logs for " + service)
	u.muted("(Log streaming API not yet connected - showing placeholder)")
	logs := []struct {
		line  string
		level string
	}{
		{"2026-05-30 16:45:30 [INFO] Service started successfully", "OK"},
		{"2026-05-30 16:45:35 [INFO] Database connection established", "OK"},
		{"2026-05-30 16:45:40 [INFO] API server listening on port 8000", "OK"},
		{"2026-05-30 16:45:45 [INFO] Health check passed", "OK"},
		{"2026-05-30 16:45:50 [WARN] High memory usage detected", "WARNING"},
		{"2026-05-30 16:45:55 [ERROR] Database connection timeout", "ERROR"},
		{"2026-05-30 16:46:00 [INFO] Database connection restored", "OK"},
	}
	for _, l := range logs {
		fmt.Println(u.colorizer.Colorize(l.line, statusStyle(l.level)))
	}
	return nil
}
