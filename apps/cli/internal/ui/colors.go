package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Color palette using lipgloss
var (
	// Primary colors
	primaryColor = lipgloss.Color("#4ECDC4") // Teal
	accentColor  = lipgloss.Color("#FF6B6B") // Red
	mutedColor   = lipgloss.Color("#999999") // Gray
	whiteColor   = lipgloss.Color("#FFFFFF") // White

	// Semantic colors
	successColor = lipgloss.Color("#4ECDC4") // Same as primary
	warningColor = lipgloss.Color("#FFA726") // Orange
	errorColor   = lipgloss.Color("#FF6B6B") // Same as accent
	infoColor    = lipgloss.Color("#42A5F5") // Blue

	// Additional colors
	yellowColor = lipgloss.Color("#FFD93D") // Yellow
)

// Base styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			Background(primaryColor).
			Padding(0, 1)

	subHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			Background(accentColor).
			Padding(0, 1)

	// Status styles
	statusSuccessStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	statusWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	statusInfoStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true)

	// Service status styles
	statusRunningStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	statusPendingStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	statusUnknownStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	// Subscriber status styles
	statusActiveStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	statusInactiveStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	statusSuspendedStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	// Metric styles
	metricGoodStyle = lipgloss.NewStyle().
			Foreground(successColor)

	metricWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	metricCriticalStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Command styles
	commandStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	optionStyle = lipgloss.NewStyle().
			Foreground(yellowColor).
			Bold(true)

	argumentStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Background(lipgloss.Color("#2D2D2D")).
				Padding(0, 1)

	tableBorderStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Faint(true)

	tableRowEvenStyle = lipgloss.NewStyle().
				Foreground(whiteColor)

	tableRowOddStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F0F0F0"))

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1)

	panelFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(1)

	// Debug and muted styles
	debugStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Faint(true)
)

// Style represents a lipgloss style
type Style struct {
	Style lipgloss.Style
}

// Predefined styles using lipgloss
var (
	StyleSuccess = Style{Style: statusSuccessStyle}
	StyleError   = Style{Style: statusErrorStyle}
	StyleWarning = Style{Style: statusWarningStyle}
	StyleInfo    = Style{Style: statusInfoStyle}
	StyleDebug   = Style{Style: debugStyle}
	StyleMuted   = Style{Style: mutedStyle}

	StyleRunning = Style{Style: statusRunningStyle}
	StyleStopped = Style{Style: statusStoppedStyle}
	StylePending = Style{Style: statusPendingStyle}
	StyleUnknown = Style{Style: statusUnknownStyle}

	StyleActive    = Style{Style: statusActiveStyle}
	StyleInactive  = Style{Style: statusInactiveStyle}
	StyleSuspended = Style{Style: statusSuspendedStyle}

	StyleHeader    = Style{Style: headerStyle}
	StyleSubHeader = Style{Style: subHeaderStyle}
	StyleTitle     = Style{Style: titleStyle}

	StyleHeaderRow = Style{Style: tableHeaderStyle}
	StyleRowEven   = Style{Style: tableRowEvenStyle}
	StyleRowOdd    = Style{Style: tableRowOddStyle}
	StyleBorder    = Style{Style: tableBorderStyle}

	StyleMetricGood     = Style{Style: metricGoodStyle}
	StyleMetricWarning  = Style{Style: metricWarningStyle}
	StyleMetricCritical = Style{Style: metricCriticalStyle}

	StyleCommand     = Style{Style: commandStyle}
	StyleOption      = Style{Style: optionStyle}
	StyleArgument    = Style{Style: argumentStyle}
	StyleDescription = Style{Style: descriptionStyle}

	StylePanel        = Style{Style: panelStyle}
	StylePanelFocused = Style{Style: panelFocusedStyle}
)

// Colorizer handles color output using lipgloss
type Colorizer struct {
	enabled bool
}

// NewColorizer creates a new colorizer
func NewColorizer(enabled bool) *Colorizer {
	return &Colorizer{enabled: enabled}
}

// IsEnabled returns whether colors are enabled
func (c *Colorizer) IsEnabled() bool {
	return c.enabled
}

// Colorize applies a style to text using lipgloss
func (c *Colorizer) Colorize(text string, style Style) string {
	if !c.enabled {
		return text
	}
	return style.Style.Render(text)
}

// Convenience methods for common colorizations
func (c *Colorizer) Success(text string) string {
	return c.Colorize(text, StyleSuccess)
}

func (c *Colorizer) Error(text string) string {
	return c.Colorize(text, StyleError)
}

func (c *Colorizer) Warning(text string) string {
	return c.Colorize(text, StyleWarning)
}

func (c *Colorizer) Info(text string) string {
	return c.Colorize(text, StyleInfo)
}

func (c *Colorizer) Debug(text string) string {
	return c.Colorize(text, StyleDebug)
}

func (c *Colorizer) Muted(text string) string {
	return c.Colorize(text, StyleMuted)
}

func (c *Colorizer) Header(text string) string {
	return c.Colorize(text, StyleHeader)
}

func (c *Colorizer) SubHeader(text string) string {
	return c.Colorize(text, StyleSubHeader)
}

func (c *Colorizer) Title(text string) string {
	return c.Colorize(text, StyleTitle)
}

func (c *Colorizer) Command(text string) string {
	return c.Colorize(text, StyleCommand)
}

func (c *Colorizer) Option(text string) string {
	return c.Colorize(text, StyleOption)
}

func (c *Colorizer) Argument(text string) string {
	return c.Colorize(text, StyleArgument)
}

func (c *Colorizer) Description(text string) string {
	return c.Colorize(text, StyleDescription)
}

// Status colorizations
func (c *Colorizer) StatusRunning(text string) string {
	return c.Colorize(text, StyleRunning)
}

func (c *Colorizer) StatusStopped(text string) string {
	return c.Colorize(text, StyleStopped)
}

func (c *Colorizer) StatusPending(text string) string {
	return c.Colorize(text, StylePending)
}

func (c *Colorizer) StatusUnknown(text string) string {
	return c.Colorize(text, StyleUnknown)
}

func (c *Colorizer) StatusActive(text string) string {
	return c.Colorize(text, StyleActive)
}

func (c *Colorizer) StatusInactive(text string) string {
	return c.Colorize(text, StyleInactive)
}

func (c *Colorizer) StatusSuspended(text string) string {
	return c.Colorize(text, StyleSuspended)
}

// Metric colorizations based on value
func (c *Colorizer) Metric(value float64, thresholds ...float64) string {
	text := fmt.Sprintf("%.1f%%", value)

	// Default thresholds if not provided
	if len(thresholds) == 0 {
		thresholds = []float64{70, 90}
	}

	switch {
	case value >= thresholds[1]:
		return c.Colorize(text, StyleMetricCritical)
	case value >= thresholds[0]:
		return c.Colorize(text, StyleMetricWarning)
	default:
		return c.Colorize(text, StyleMetricGood)
	}
}

// Panel rendering
func (c *Colorizer) Panel(text string, focused bool) string {
	if focused {
		return c.Colorize(text, StylePanelFocused)
	}
	return c.Colorize(text, StylePanel)
}

// Check if colors should be disabled
func ShouldDisableColors() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return true
	}

	// Check if output is not a TTY
	if fileInfo, _ := os.Stdout.Stat(); fileInfo != nil {
		return (fileInfo.Mode() & 0o100) == 0 // Not a TTY
	}

	return false
}

// Auto-detect color support
func AutoDetectColors() bool {
	return !ShouldDisableColors()
}

// Global colorizer instance
var GlobalColorizer *Colorizer

// InitColors initializes the global colorizer
func InitColors(enabled bool) {
	GlobalColorizer = NewColorizer(enabled)
}
