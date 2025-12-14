package components

import "github.com/charmbracelet/lipgloss"

// Design tokens for UI components (centralized here to avoid import cycles).
var (
	ColorBackground = lipgloss.Color("#000000")
	ColorPrimary    = lipgloss.Color("#69F0AE")
	ColorAccent     = lipgloss.Color("#FFD54F")
	ColorDanger     = lipgloss.Color("#FF6B6B")
	ColorInfo       = lipgloss.Color("#4DD0E1")
	ColorPurple     = lipgloss.Color("#CE93D8")
	ColorOrange     = lipgloss.Color("#FFA726")
	ColorMuted      = lipgloss.Color("#B0B0B0")
	ColorBoxDark    = lipgloss.Color("#071013")
	ColorBorderDeep = lipgloss.Color("#1F3B2E")

	PaddingSmall  = 1
	PaddingMedium = 2
	PaddingLarge  = 4
)
