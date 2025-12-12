package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/ui/components"
)

var (
	equipmentStyle      = lipgloss.NewStyle().Padding(1, 2)
	equipmentTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	equipmentItemStyle  = lipgloss.NewStyle().PaddingLeft(2)
)

// EquipmentModel displays and manages player equipment.
type EquipmentModel struct {
	playerStats game.Stats
	equipment   map[string]string // slot -> equipment name
	cursor      int
	slots       []string
}

// NewEquipmentModel creates a new EquipmentModel.
func NewEquipmentModel(stats game.Stats) EquipmentModel {
	// Sample equipment (in real implementation, fetch from DB)
	equipment := map[string]string{
		"weapon": "Wooden Sword",
		"armor":  "Leather Armor",
		"ring":   "None",
		"charm":  "None",
	}
	slots := []string{"weapon", "armor", "ring", "charm"}
	return EquipmentModel{
		playerStats: stats,
		equipment:   equipment,
		cursor:      0,
		slots:       slots,
	}
}

func (m EquipmentModel) Init() tea.Cmd {
	return nil
}

func (m EquipmentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return EquipmentToTownMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.slots)-1 {
				m.cursor++
			}
		case "enter":
			// TODO: Open equipment selection for the slot
			// For now, just cycle through sample equipment
			slot := m.slots[m.cursor]
			switch m.equipment[slot] {
			case "Wooden Sword":
				m.equipment[slot] = "Iron Sword"
			case "Iron Sword":
				m.equipment[slot] = "Steel Sword"
			case "Steel Sword":
				m.equipment[slot] = "Wooden Sword"
			case "Leather Armor":
				m.equipment[slot] = "Chain Mail"
			case "Chain Mail":
				m.equipment[slot] = "Plate Armor"
			case "Plate Armor":
				m.equipment[slot] = "Leather Armor"
			case "None":
				m.equipment[slot] = "Basic Ring"
			case "Basic Ring":
				m.equipment[slot] = "None"
			}
		}
	}
	return m, nil
}

func (m EquipmentModel) View() string {
	s := m.playerStats
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(lipgloss.Width(components.View(s))).Render("TUI English Quest"),
		components.View(s),
	)

	var b strings.Builder
	b.WriteString(equipmentTitleStyle.Render("Equipment\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	b.WriteString("Select a slot and press Enter to change equipment.\n\n")

	for i, slot := range m.slots {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		equip := m.equipment[slot]
		effect := " (No effect)" // TODO: Calculate actual effects
		line := fmt.Sprintf("%s%-8s: %s%s\n", cursor, strings.Title(slot), equip, effect)
		b.WriteString(line)
	}

	footer := "\n[j/k] Navigate  [Enter] Change  [Esc] Back to Town"

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		equipmentStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
