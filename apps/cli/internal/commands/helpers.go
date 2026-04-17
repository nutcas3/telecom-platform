package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/api"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/ui"
)

// uiContext bundles the common UI primitives used across commands.
type uiContext struct {
	colorizer *ui.Colorizer
	icons     *ui.IconRenderer
	client    *api.Client
	config    *types.CLIConfig
	connected bool
}

// newUIContext builds a uiContext from a CLIConfig, probing API connectivity.
func newUIContext(cfg *types.CLIConfig) *uiContext {
	useColor := cfg == nil || !cfg.NoColor
	client := api.NewClient(cfg)
	return &uiContext{
		colorizer: ui.NewColorizer(useColor),
		icons:     ui.NewIconRenderer(true, false),
		client:    client,
		config:    cfg,
		connected: client.IsConnected(),
	}
}

// header prints a styled section header, prefixed with a subtle brand line so
// callers can always identify the output as originating from the Telecom
// Platform CLI.
func (u *uiContext) header(title string) {
	fmt.Println(u.colorizer.Colorize("Telecom Platform CLI", ui.StyleMuted))
	fmt.Println(u.colorizer.Colorize(title, ui.StyleHeader))
}

// info prints a styled informational line.
func (u *uiContext) info(msg string) {
	fmt.Println(u.colorizer.Colorize(msg, ui.StyleInfo))
}

// success prints a styled success line.
func (u *uiContext) success(msg string) {
	fmt.Println(u.colorizer.Colorize(msg, ui.StyleSuccess))
}

// warn prints a styled warning line.
func (u *uiContext) warn(msg string) {
	fmt.Println(u.colorizer.Colorize(msg, ui.StyleWarning))
}

// error prints a styled error line.
func (u *uiContext) errorln(msg string) {
	fmt.Println(u.colorizer.Colorize(msg, ui.StyleError))
}

// muted prints a styled muted line.
func (u *uiContext) muted(msg string) {
	fmt.Println(u.colorizer.Colorize(msg, ui.StyleMuted))
}

// connectivityBanner prints a warning banner if the API is unreachable.
func (u *uiContext) connectivityBanner() {
	if !u.connected {
		u.warn("Warning: API server unreachable at " + u.endpointDisplay() + " - showing placeholder data")
		fmt.Println()
	}
}

func (u *uiContext) endpointDisplay() string {
	if u.config == nil || u.config.APIEndpoint == "" {
		return "(no endpoint)"
	}
	return u.config.APIEndpoint
}

// newTable returns a freshly configured UI table.
func (u *uiContext) newTable() *ui.Table {
	return ui.NewTable(u.colorizer, u.icons)
}

// statusStyle maps a status string to a ui.Style.
func statusStyle(status string) ui.Style {
	switch status {
	case "PAID", "SUCCESS", "COMPLETED", "HEALTHY", "ACTIVE", "RUNNING", "OK":
		return ui.StyleSuccess
	case "PENDING", "PROCESSING", "BUILDING", "WARNING", "WARN":
		return ui.StyleWarning
	case "FAILED", "CANCELLED", "OVERDUE", "INACTIVE", "STOPPED", "ERROR", "CRITICAL":
		return ui.StyleError
	default:
		return ui.StyleMuted
	}
}
