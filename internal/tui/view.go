package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m *Model) View() string {
	cmdDisplay := m.command.Value()
	if cmdDisplay == "" {
		cmdDisplay = "not set"
	}
	cmdLine := inactiveStyle.Render(fmt.Sprintf("cmd: %s", cmdDisplay))
	cmdTop := lipgloss.Place(m.w, 1, lipgloss.Center, lipgloss.Top, cmdLine)

	var content string

	if m.editingCommand {
		content = lipgloss.JoinVertical(lipgloss.Center, m.command.View())
	} else {
		content = lipgloss.JoinVertical(lipgloss.Center,
			m.username.View(),
			m.password.View(),
		)
	}

	if m.errorMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			content,
			"",
			errorStyle.Render(m.errorMsg),
		)
	}

	var help string
	if m.editingCommand {
		help = inactiveStyle.Render("Enter confirm • ESC cancel")
	} else {
		help = inactiveStyle.Render("F2 edit command")
	}

	main := lipgloss.Place(m.w, m.h-3, lipgloss.Center, lipgloss.Center, content)
	helpBar := lipgloss.Place(m.w, 1, lipgloss.Center, lipgloss.Top, help)

	return lipgloss.JoinVertical(lipgloss.Left, cmdTop, main, helpBar)
}
