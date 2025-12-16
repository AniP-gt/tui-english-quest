package components

import (
	"fmt"
	"strings"
)

// RenderKeyValue renders a left-aligned key-value pair with a fixed label width.
func RenderKeyValue(label string, value string, labelWidth int) string {
	return fmt.Sprintf("%-*s %s", labelWidth, label, value)
}

// RenderBulletList renders a list of items as a bulleted list with a specified indent.
func RenderBulletList(items []string, indent int) string {
	var b strings.Builder
	prefix := strings.Repeat(" ", indent) + "- "
	for _, item := range items {
		b.WriteString(prefix + item + "\n")
	}
	return b.String()
}

// RenderAlignedRow renders a row of columns with specified widths.
func RenderAlignedRow(cols []string, widths []int) string {
	var b strings.Builder
	for i, col := range cols {
		width := widths[i]
		// Use a negative width for left alignment in fmt.Sprintf
		b.WriteString(fmt.Sprintf("%-*s", width, col))
	}
	return b.String()
}
