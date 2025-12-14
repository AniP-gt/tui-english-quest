package components

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// Menu renders items in columns with a selected index.
func Menu(items []string, selected int, cols int, width int) string {
	if cols <= 0 {
		cols = 2
	}
	var columns [][]string
	for i := 0; i < cols; i++ {
		columns = append(columns, []string{})
	}
	for i, it := range items {
		col := i % cols
		if i == selected {
			columns[col] = append(columns[col], lipgloss.NewStyle().Background(ColorAccent).Foreground(ColorBoxDark).Padding(0, 1).Render("> "+it))
		} else {
			columns[col] = append(columns[col], lipgloss.NewStyle().Foreground(ColorPrimary).Padding(0, 1).Render("  "+it))
		}
	}
	// Join each column vertically
	var colStrs []string
	for _, col := range columns {
		colStrs = append(colStrs, strings.Join(col, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, colStrs...)
}
