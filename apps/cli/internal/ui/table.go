package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TableColumn represents a column in a table
type TableColumn struct {
	Header      string
	Width       int
	Alignment   string // "left", "right", "center"
	Style       lipgloss.Style
	HeaderStyle lipgloss.Style
}

// TableRow represents a row in a table
type TableRow struct {
	Cells []TableCell
	Style lipgloss.Style
}

// TableCell represents a cell in a table
type TableCell struct {
	Text  string
	Style lipgloss.Style
	Align string
}

// Table represents a formatted table using lipgloss
type Table struct {
	Columns      []TableColumn
	Rows         []TableRow
	Style        TableStyle
	Colorizer    *Colorizer
	IconRenderer *IconRenderer
}

// TableStyle defines the visual style of a table
type TableStyle struct {
	Border          bool
	HeaderSeparator bool
	RowSeparator    bool
	Padding         int
	BorderColor     lipgloss.Style
	HeaderColor     lipgloss.Style
	EvenRowColor    lipgloss.Style
	OddRowColor     lipgloss.Style
}

// NewTable creates a new table with lipgloss styling
func NewTable(colorizer *Colorizer, iconRenderer *IconRenderer) *Table {
	return &Table{
		Colorizer:    colorizer,
		IconRenderer: iconRenderer,
		Style: TableStyle{
			Border:          true,
			HeaderSeparator: true,
			RowSeparator:    false,
			Padding:         1,
			BorderColor:     tableBorderStyle,
			HeaderColor:     tableHeaderStyle,
			EvenRowColor:    tableRowEvenStyle,
			OddRowColor:     tableRowOddStyle,
		},
	}
}

// AddColumn adds a column to the table
func (t *Table) AddColumn(header string, width int, alignment string) {
	style := lipgloss.NewStyle()
	headerStyle := tableHeaderStyle

	t.Columns = append(t.Columns, TableColumn{
		Header:      header,
		Width:       width,
		Alignment:   alignment,
		Style:       style,
		HeaderStyle: headerStyle,
	})
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	row := TableRow{Style: lipgloss.NewStyle()}

	for _, cell := range cells {
		row.Cells = append(row.Cells, TableCell{
			Text:  cell,
			Style: lipgloss.NewStyle(),
			Align: "left",
		})
	}

	t.Rows = append(t.Rows, row)
}

// AddStyledRow adds a styled row to the table
func (t *Table) AddStyledRow(style lipgloss.Style, cells ...string) {
	row := TableRow{Style: style}

	for _, cell := range cells {
		row.Cells = append(row.Cells, TableCell{
			Text:  cell,
			Style: lipgloss.NewStyle(),
			Align: "left",
		})
	}

	t.Rows = append(t.Rows, row)
}

// AddStyledCellRow adds a row with individually styled cells
func (t *Table) AddStyledCellRow(cells ...TableCell) {
	row := TableRow{Style: lipgloss.NewStyle()}
	row.Cells = cells
	t.Rows = append(t.Rows, row)
}

// Render renders the table as a string using lipgloss
func (t *Table) Render() string {
	if len(t.Columns) == 0 {
		return ""
	}

	var rows []string

	// Render header
	if t.Style.Border {
		rows = append(rows, t.renderBorderLine("top"))
	}

	headerRow := t.renderHeader()
	rows = append(rows, headerRow)

	if t.Style.HeaderSeparator {
		rows = append(rows, t.renderBorderLine("header"))
	}

	// Render rows
	for i, row := range t.Rows {
		if t.Style.RowSeparator && i > 0 {
			rows = append(rows, t.renderBorderLine("row"))
		}

		rowStyle := t.Style.EvenRowColor
		if i%2 == 1 {
			rowStyle = t.Style.OddRowColor
		}

		renderedRow := t.renderRow(row, rowStyle)
		rows = append(rows, renderedRow)
	}

	// Render bottom border
	if t.Style.Border {
		rows = append(rows, t.renderBorderLine("bottom"))
	}

	return strings.Join(rows, "\n")
}

// renderHeader renders the table header
func (t *Table) renderHeader() string {
	var cells []string

	for _, col := range t.Columns {
		paddedHeader := t.padText(col.Header, col.Width, col.Alignment)
		styledHeader := col.HeaderStyle.Render(paddedHeader)
		cells = append(cells, styledHeader)
	}

	return t.renderRowLine(cells, t.Style.BorderColor)
}

// renderRow renders a table row
func (t *Table) renderRow(row TableRow, defaultStyle lipgloss.Style) string {
	var cells []string

	for i, cell := range row.Cells {
		colStyle := defaultStyle
		if i < len(t.Columns) {
			colStyle = t.Columns[i].Style
		}

		cellStyle := cell.Style
		if cellStyle.String() == lipgloss.NewStyle().String() {
			cellStyle = colStyle
		}

		align := cell.Align
		if align == "" && i < len(t.Columns) {
			align = t.Columns[i].Alignment
		}

		width := 20
		if i < len(t.Columns) {
			width = t.Columns[i].Width
		}

		paddedText := t.padText(cell.Text, width, align)
		styledText := cellStyle.Render(paddedText)
		cells = append(cells, styledText)
	}

	return t.renderRowLine(cells, defaultStyle)
}

// renderRowLine renders a line of cells
func (t *Table) renderRowLine(cells []string, borderStyle lipgloss.Style) string {
	if t.Style.Border {
		border := borderStyle.Render(" ")
		separator := borderStyle.Render(" ")
		return border + " " + strings.Join(cells, " "+separator+" ") + " " + border
	} else {
		return strings.Join(cells, " ")
	}
}

// renderBorderLine renders a border line
func (t *Table) renderBorderLine(position string) string {
	if !t.Style.Border {
		return ""
	}

	var sections []string
	var borderChar string

	// Use position parameter to determine border style
	switch position {
	case "top", "bottom":
		borderChar = "="
	case "middle":
		borderChar = "-"
	default:
		borderChar = "="
	}

	for _, col := range t.Columns {
		width := col.Width + 2 // Account for padding
		if t.Style.Padding > 0 {
			width += t.Style.Padding * 2
		}
		sections = append(sections, strings.Repeat(borderChar, width))
	}

	border := t.Style.BorderColor.Render("+")
	separator := t.Style.BorderColor.Render("+")

	return border + strings.Join(sections, separator) + border
}

// padText pads text to the specified width with alignment
func (t *Table) padText(text string, width int, alignment string) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := width - len(text)

	switch alignment {
	case "right":
		return strings.Repeat(" ", padding) + text
	case "center":
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default: // left
		return text + strings.Repeat(" ", padding)
	}
}

// NewServicesTable creates a table optimized for service listings with lipgloss styling
func NewServicesTable(colorizer *Colorizer, iconRenderer *IconRenderer) *Table {
	table := NewTable(colorizer, iconRenderer)
	table.AddColumn("", 3, "center") // Icon
	table.AddColumn("Service", 15, "left")
	table.AddColumn("Status", 10, "center")
	table.AddColumn("Version", 10, "center")
	table.AddColumn("CPU", 6, "right")
	table.AddColumn("Memory", 8, "right")
	table.AddColumn("Uptime", 8, "right")
	return table
}

// NewSubscribersTable creates a table optimized for subscriber listings with lipgloss styling
func NewSubscribersTable(colorizer *Colorizer, iconRenderer *IconRenderer) *Table {
	table := NewTable(colorizer, iconRenderer)
	table.AddColumn("", 3, "center") // Icon
	table.AddColumn("IMSI", 15, "left")
	table.AddColumn("Name", 18, "left")
	table.AddColumn("Status", 10, "center")
	table.AddColumn("Balance", 10, "right")
	table.AddColumn("Plan", 10, "left")
	table.AddColumn("Last Activity", 12, "center")
	return table
}

// NewMetricsTable creates a table optimized for metrics display with lipgloss styling
func NewMetricsTable(colorizer *Colorizer, iconRenderer *IconRenderer) *Table {
	table := NewTable(colorizer, iconRenderer)
	table.AddColumn("", 3, "center") // Icon
	table.AddColumn("Metric", 20, "left")
	table.AddColumn("Value", 12, "right")
	table.AddColumn("Status", 10, "center")
	table.AddColumn("Threshold", 12, "right")
	return table
}

// NewSimpleTable creates a simple table with basic columns and lipgloss styling
func NewSimpleTable(colorizer *Colorizer, iconRenderer *IconRenderer, headers ...string) *Table {
	table := NewTable(colorizer, iconRenderer)

	for _, header := range headers {
		table.AddColumn(header, 15, "left")
	}

	return table
}

// NewStyledTable creates a table with custom styling
func NewStyledTable(colorizer *Colorizer, iconRenderer *IconRenderer, style TableStyle) *Table {
	table := &Table{
		Colorizer:    colorizer,
		IconRenderer: iconRenderer,
		Style:        style,
	}
	return table
}
