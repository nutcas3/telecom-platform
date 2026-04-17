package ui

// Icons and symbols for CLI output
const (
	// Status icons
	IconSuccess = "check"
	IconError   = "x"
	IconWarning = "warning"
	IconInfo    = "info"
	IconDebug   = "debug"
	IconLoading = "dots"

	// Service status icons
	IconRunning = "play"
	IconStopped = "stop"
	IconPending = "clock"
	IconUnknown = "help"

	// Subscriber status icons
	IconActive    = "check-circle"
	IconInactive  = "x-circle"
	IconSuspended = "pause-circle"

	// Resource icons
	IconCPU     = "cpu"
	IconMemory  = "hard-drive"
	IconDisk    = "database"
	IconNetwork = "globe"
	IconAPI     = "api"
	IconKey     = "key"
	IconLock    = "lock"

	// Navigation icons
	IconArrowRight = "arrow-right"
	IconArrowDown  = "arrow-down"
	IconFolder     = "folder"
	IconFile       = "file"
	IconHome       = "home"

	// Action icons
	IconAdd      = "plus"
	IconRemove   = "minus"
	IconEdit     = "edit"
	IconDelete   = "trash"
	IconRefresh  = "refresh-cw"
	IconDownload = "download"
	IconUpload   = "upload"

	// Communication icons
	IconMail    = "mail"
	IconPhone   = "phone"
	IconMessage = "message-square"
	IconBell    = "bell"

	// Security icons
	IconShield = "shield"
	IconEye    = "eye"
	IconEyeOff = "eye-off"

	// Chart icons
	IconChart     = "trending-up"
	IconChartDown = "trending-down"
	IconBarChart  = "bar-chart"
	IconPieChart  = "pie-chart"

	// Container/platform icons
	IconContainer = "box"
	IconK8s       = "kubernetes"
	IconDocker    = "docker"
	IconCloud     = "cloud"
	IconServer    = "server"
)

// Unicode fallback icons for terminals without emoji support
var UnicodeIcons = map[string]string{
	IconSuccess:    "OK",
	IconError:      "ERR",
	IconWarning:    "WARN",
	IconInfo:       "INFO",
	IconDebug:      "DEBUG",
	IconLoading:    "...",
	IconRunning:    "RUN",
	IconStopped:    "STOP",
	IconPending:    "PEND",
	IconUnknown:    "UNK",
	IconActive:     "ACT",
	IconInactive:   "INACT",
	IconSuspended:  "SUSP",
	IconCPU:        "CPU",
	IconMemory:     "MEM",
	IconDisk:       "DISK",
	IconNetwork:    "NET",
	IconAPI:        "API",
	IconKey:        "KEY",
	IconLock:       "LCK",
	IconArrowRight: "->",
	IconArrowDown:  "v",
	IconFolder:     "DIR",
	IconFile:       "FILE",
	IconHome:       "HOME",
	IconAdd:        "+",
	IconRemove:     "-",
	IconEdit:       "EDIT",
	IconDelete:     "DEL",
	IconRefresh:    "REF",
	IconDownload:   "DL",
	IconUpload:     "UL",
	IconMail:       "MAIL",
	IconPhone:      "PHONE",
	IconMessage:    "MSG",
	IconBell:       "BELL",
	IconShield:     "SHLD",
	IconEye:        "EYE",
	IconEyeOff:     "HIDE",
	IconChart:      "UP",
	IconChartDown:  "DOWN",
	IconBarChart:   "BAR",
	IconPieChart:   "PIE",
	IconContainer:  "BOX",
	IconK8s:        "K8S",
	IconDocker:     "DOCKER",
	IconCloud:      "CLOUD",
	IconServer:     "SRV",
}

// Emoji icons for modern terminals
var EmojiIcons = map[string]string{
	IconSuccess:    "OK",
	IconError:      "ERR",
	IconWarning:    "WARN",
	IconInfo:       "INFO",
	IconDebug:      "DEBUG",
	IconLoading:    "...",
	IconRunning:    "RUN",
	IconStopped:    "STOP",
	IconPending:    "PEND",
	IconUnknown:    "UNK",
	IconActive:     "ACT",
	IconInactive:   "INACT",
	IconSuspended:  "SUSP",
	IconCPU:        "CPU",
	IconMemory:     "MEM",
	IconDisk:       "DISK",
	IconNetwork:    "NET",
	IconAPI:        "API",
	IconKey:        "KEY",
	IconLock:       "LCK",
	IconArrowRight: "->",
	IconArrowDown:  "v",
	IconFolder:     "DIR",
	IconFile:       "FILE",
	IconHome:       "HOME",
	IconAdd:        "+",
	IconRemove:     "-",
	IconEdit:       "EDIT",
	IconDelete:     "DEL",
	IconRefresh:    "REF",
	IconDownload:   "DL",
	IconUpload:     "UL",
	IconMail:       "MAIL",
	IconPhone:      "PHONE",
	IconMessage:    "MSG",
	IconBell:       "BELL",
	IconShield:     "SHLD",
	IconEye:        "EYE",
	IconEyeOff:     "HIDE",
	IconChart:      "UP",
	IconChartDown:  "DOWN",
	IconBarChart:   "BAR",
	IconPieChart:   "PIE",
	IconContainer:  "BOX",
	IconK8s:        "K8S",
	IconDocker:     "DOCKER",
	IconCloud:      "CLOUD",
	IconServer:     "SRV",
}

// IconRenderer handles icon rendering
type IconRenderer struct {
	useUnicode bool
	useEmoji   bool
}

// NewIconRenderer creates a new icon renderer
func NewIconRenderer(useUnicode, useEmoji bool) *IconRenderer {
	return &IconRenderer{
		useUnicode: useUnicode,
		useEmoji:   useEmoji,
	}
}

// Render renders an icon based on the current mode
func (ir *IconRenderer) Render(iconName string) string {
	if ir.useEmoji {
		if icon, exists := EmojiIcons[iconName]; exists {
			return icon
		}
	}

	if ir.useUnicode {
		if icon, exists := UnicodeIcons[iconName]; exists {
			return icon
		}
	}

	// Fallback to text version
	if icon, exists := UnicodeIcons[iconName]; exists {
		return icon
	}

	return iconName
}

// Convenience methods for common icons
func (ir *IconRenderer) Success() string {
	return ir.Render(IconSuccess)
}

func (ir *IconRenderer) Error() string {
	return ir.Render(IconError)
}

func (ir *IconRenderer) Warning() string {
	return ir.Render(IconWarning)
}

func (ir *IconRenderer) Info() string {
	return ir.Render(IconInfo)
}

func (ir *IconRenderer) Running() string {
	return ir.Render(IconRunning)
}

func (ir *IconRenderer) Stopped() string {
	return ir.Render(IconStopped)
}

func (ir *IconRenderer) Pending() string {
	return ir.Render(IconPending)
}

func (ir *IconRenderer) Active() string {
	return ir.Render(IconActive)
}

func (ir *IconRenderer) Inactive() string {
	return ir.Render(IconInactive)
}

func (ir *IconRenderer) CPU() string {
	return ir.Render(IconCPU)
}

func (ir *IconRenderer) Memory() string {
	return ir.Render(IconMemory)
}

func (ir *IconRenderer) Network() string {
	return ir.Render(IconNetwork)
}

func (ir *IconRenderer) API() string {
	return ir.Render(IconAPI)
}

func (ir *IconRenderer) Key() string {
	return ir.Render(IconKey)
}

func (ir *IconRenderer) Shield() string {
	return ir.Render(IconShield)
}

func (ir *IconRenderer) Chart() string {
	return ir.Render(IconChart)
}

func (ir *IconRenderer) Container() string {
	return ir.Render(IconContainer)
}

func (ir *IconRenderer) Cloud() string {
	return ir.Render(IconCloud)
}

func (ir *IconRenderer) Server() string {
	return ir.Render(IconServer)
}

func (ir *IconRenderer) Clock() string {
	return ir.Render(IconPending)
}

// Status icon with color
func (ir *IconRenderer) StatusColored(status string, colorizer *Colorizer) string {
	switch status {
	case "running", "active":
		return colorizer.StatusRunning(ir.Running())
	case "stopped", "inactive":
		return colorizer.StatusStopped(ir.Stopped())
	case "pending":
		return colorizer.StatusPending(ir.Pending())
	default:
		return colorizer.StatusUnknown(status)
	}
}

// Global icon renderer
var GlobalIconRenderer *IconRenderer

// InitIcons initializes the global icon renderer
func InitIcons(useUnicode, useEmoji bool) {
	GlobalIconRenderer = NewIconRenderer(useUnicode, useEmoji)
}
