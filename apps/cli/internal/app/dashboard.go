package app

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/api"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

type Dashboard struct {
	table      table.Model
	spinner    spinner.Model
	help       help.Model
	keys       keyMap
	quitting   bool
	loading    bool
	lastUpdate time.Time
	client     *api.Client
	connected  bool
	statusMsg  string
}

type keyMap struct {
	refresh     key.Binding
	subscribers key.Binding
	services    key.Binding
	billing     key.Binding
	monitoring  key.Binding
	config      key.Binding
	deploy      key.Binding
	help        key.Binding
	quit        key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.refresh, k.subscribers, k.services},
		{k.billing, k.monitoring, k.config},
		{k.deploy, k.help, k.quit},
	}
}

var defaultKeyMap = keyMap{
	refresh:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	subscribers: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "subscribers")),
	services:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "services")),
	billing:     key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "billing")),
	monitoring:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "monitoring")),
	config:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "config")),
	deploy:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "deploy")),
	help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func NewDashboard() *Dashboard {
	return NewDashboardWithConfig(nil)
}

// NewDashboardWithConfig builds a dashboard model wired to the telecom API via
// the provided CLI configuration.
func NewDashboardWithConfig(cfg *types.CLIConfig) *Dashboard {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	t := table.New()
	t.SetColumns([]table.Column{
		{Title: "Service", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "CPU", Width: 8},
		{Title: "Memory", Width: 10},
		{Title: "Uptime", Width: 12},
		{Title: "Version", Width: 10},
	})
	t.SetStyles(table.DefaultStyles())

	client := api.NewClient(cfg)
	d := &Dashboard{
		table:      t,
		spinner:    s,
		help:       help.New(),
		keys:       defaultKeyMap,
		lastUpdate: time.Now(),
		client:     client,
		connected:  client.IsConnected(),
	}
	d.populateRows()
	return d
}

// populateRows fetches live service data from the API and falls back to
// placeholder rows if the API is unreachable or empty.
func (d *Dashboard) populateRows() {
	services, err := d.client.ListServices()
	if err != nil || len(services) == 0 {
		if err != nil {
			d.statusMsg = "API unreachable - showing sample data"
		} else {
			d.statusMsg = "API returned no services - showing sample data"
		}
		d.table.SetRows([]table.Row{
			{"API Server", "Running", "45%", "256MB", "2h15m", "v1.0.0"},
			{"Charging Engine", "Running", "23%", "128MB", "2h15m", "v1.0.0"},
			{"Packet Gateway", "Running", "67%", "512MB", "2h15m", "v1.0.0"},
			{"Web Dashboard", "Running", "12%", "64MB", "2h15m", "v1.0.0"},
			{"Prometheus", "Running", "8%", "32MB", "2h15m", "v2.45.0"},
			{"Grafana", "Running", "5%", "24MB", "2h15m", "v9.5.2"},
		})
		return
	}

	d.statusMsg = fmt.Sprintf("Connected to %s", clientEndpoint(d.client))
	rows := make([]table.Row, 0, len(services))
	for _, s := range services {
		rows = append(rows, table.Row{
			s.Name,
			s.Status,
			fmt.Sprintf("%.1f%%", s.CPU),
			s.Memory,
			s.Uptime,
			s.Version,
		})
	}
	d.table.SetRows(rows)
}

// clientEndpoint is a tiny accessor used only for the dashboard status line.
func clientEndpoint(c *api.Client) string {
	if c == nil {
		return "(no client)"
	}
	return c.BaseURL()
}

func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(d.spinner.Tick)
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.quit):
			d.quitting = true
			return d, tea.Quit
		case key.Matches(msg, d.keys.help):
			d.help.ShowAll = !d.help.ShowAll
			return d, nil
		case key.Matches(msg, d.keys.refresh):
			d.loading = true
			return d, d.refreshData
		case key.Matches(msg, d.keys.subscribers):
			// Navigate to subscribers view
			return d, tea.Quit
		case key.Matches(msg, d.keys.services):
			// Navigate to services view
			return d, tea.Quit
		case key.Matches(msg, d.keys.billing):
			// Navigate to billing view
			return d, tea.Quit
		case key.Matches(msg, d.keys.monitoring):
			// Navigate to monitoring view
			return d, tea.Quit
		case key.Matches(msg, d.keys.config):
			// Navigate to config view
			return d, tea.Quit
		case key.Matches(msg, d.keys.deploy):
			// Navigate to deploy view
			return d, tea.Quit
		}
	// Timer removed - manual refresh only
	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	}

	d.table, cmd = d.table.Update(msg)
	return d, cmd
}

func (d *Dashboard) View() string {
	title := d.titleView()
	table := d.table.View()
	footer := d.footerView()

	helpView := ""
	if d.help.ShowAll {
		helpView = d.help.View(d.keys)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s",
		title,
		table,
		footer,
		helpView,
	)
}

func (d *Dashboard) titleView() string {
	title := "Telecom Platform Dashboard"
	subtitle := fmt.Sprintf("Last updated: %s", d.lastUpdate.Format("15:04:05"))

	if d.loading {
		subtitle += " " + d.spinner.View()
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("240")).
		Padding(0, 2).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2)

	return titleStyle.Render(title) + "\n" + subtitleStyle.Render(subtitle)
}

func (d *Dashboard) footerView() string {
	info := "Press '?' for help | 'r' to refresh | 'q' to quit"

	totalServices := len(d.table.Rows())
	runningServices := 0
	for _, row := range d.table.Rows() {
		if len(row) > 1 && row[1] == "Running" {
			runningServices++
		}
	}

	connectionStatus := "disconnected"
	connectionColor := lipgloss.Color("196")
	if d.connected {
		connectionStatus = "connected"
		connectionColor = lipgloss.Color("46")
	}
	connectionStyle := lipgloss.NewStyle().Foreground(connectionColor).Bold(true)
	connection := fmt.Sprintf("API: %s (%s)", connectionStyle.Render(connectionStatus), clientEndpoint(d.client))

	status := fmt.Sprintf("Services: %d/%d running", runningServices, totalServices)
	statusLine := ""
	if d.statusMsg != "" {
		statusLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).Padding(0, 2).Render(d.statusMsg)
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2)

	return footerStyle.Render(fmt.Sprintf("%s | %s | %s", info, status, connection)) + statusLine
}

func (d *Dashboard) refreshData() tea.Msg {
	d.loading = false
	d.lastUpdate = time.Now()
	d.connected = d.client.IsConnected()
	d.populateRows()
	return nil
}

// RunDashboard launches the interactive dashboard with an optional config so
// the dashboard can connect to the correct API endpoint.
func RunDashboard(cfg *types.CLIConfig) {
	p := tea.NewProgram(NewDashboardWithConfig(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running dashboard: %v\n", err)
		os.Exit(1)
	}
}
