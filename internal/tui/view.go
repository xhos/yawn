package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	inactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m *Model) View() string {
	var sections []string

	// top command display
	if !m.minimal {
		cmdDisplay := m.command.Value()
		if cmdDisplay == "" {
			cmdDisplay = "not set"
		}
		cmdLine := inactiveStyle.Render(fmt.Sprintf("cmd: %s", cmdDisplay))
		sections = append(sections, lipgloss.PlaceHorizontal(m.w, lipgloss.Center, cmdLine))
	}

	// main content
	var contentParts []string
	if m.editingCommand {
		contentParts = []string{m.command.View()}
	} else {
		contentParts = []string{m.username.View(), m.password.View()}
	}

	// reserve space for the error
	contentParts = append(contentParts, "")
	if m.errorMsg != "" {
		contentParts = append(contentParts, errorStyle.Render(m.errorMsg))
	} else {
		contentParts = append(contentParts, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Center, contentParts...)

	// fix height based on minimal mode
	mainHeight := m.h
	if !m.minimal {
		mainHeight = m.h - 2
	}
	main := lipgloss.Place(m.w, mainHeight, lipgloss.Center, lipgloss.Center, content)
	sections = append(sections, main)

	// bottom help display
	if !m.minimal {
		var helpDisplay string
		if m.editingCommand {
			helpDisplay = "Enter confirm • ESC cancel"
		} else {
			helpDisplay = "F2 edit command"
		}
		helpBar := lipgloss.PlaceHorizontal(m.w, lipgloss.Center, inactiveStyle.Render(helpDisplay))
		sections = append(sections, helpBar)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
