package tui

import (
	"fmt"
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xhos/yawn/internal/greetd"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelAndClose()
			return m, tea.Quit

		case "esc":
			if m.editingCommand {
				m.editingCommand = false
				m.focused = focusPassword
				m.updateFocus()
				return m, nil
			}
			m.cancelAndClose()
			return m, tea.Quit

		case "f2":
			if !m.waitingForPAM {
				m.editingCommand = !m.editingCommand
				if m.editingCommand {
					m.focused = focusCommand
				} else {
					m.focused = focusPassword
				}
				m.updateFocus()
			}
			return m, nil

		case "tab", "shift+tab":
			if !m.waitingForPAM && !m.editingCommand {
				m.focused = (m.focused + 1) % 2
				m.updateFocus()
			}
			return m, nil

		case "enter":
			m.errorMsg = ""

			if m.editingCommand {
				m.editingCommand = false
				m.focused = focusPassword
				m.updateFocus()
				return m, nil
			}

			if m.waitingForPAM {
				return m, nil
			}

			if m.username.Value() == "" {
				m.focused = focusUsername
				m.updateFocus()
				return m, nil
			}

			if m.password.Value() == "" {
				m.focused = focusPassword
				m.updateFocus()
				return m, nil
			}

			if m.command.Value() == "" {
				return m, m.setError("command not set (press F2 to edit)")
			}

			if m.conn != nil {
				m.conn.Close()
				m.conn = nil
			}

			conn, err := net.Dial("unix", m.sockPath)
			if err != nil {
				return m, m.setError(fmt.Sprintf("failed to connect: %s", err))
			}
			m.conn = conn

			m.waitingForPAM = true
			m.starting = false

			req := &greetd.Request{
				Type:     "create_session",
				Username: m.username.Value(),
			}
			if err := req.Encode(m.conn); err != nil {
				m.waitingForPAM = false
				m.conn.Close()
				m.conn = nil
				return m, m.setError(fmt.Sprintf("failed to create session: %s", err))
			}

			return m, m.waitForResponse()
		}

	case greetdResponseMsg:
		m.waitingForPAM = false

		if msg.err != nil {
			m.cancelAndClose()
			m.starting = false
			m.password.SetValue("")
			return m, m.setError(fmt.Sprintf("connection error: %s", msg.err))
		}

		switch msg.resp.Type {
		case "success":
			if m.starting {
				return m, tea.Quit
			}

			m.starting = true
			req := &greetd.Request{
				Type: "start_session",
				Cmd:  []string{"sh", "-c", m.command.Value()},
				Env:  []string{},
			}
			if err := req.Encode(m.conn); err != nil {
				m.cancelAndClose()
				m.starting = false
				m.password.SetValue("")
				return m, m.setError(fmt.Sprintf("failed to start session: %s", err))
			}

			m.waitingForPAM = true
			return m, m.waitForResponse()

		case "error":
			m.cancelAndClose()
			m.starting = false
			m.password.SetValue("")

			if msg.resp.ErrorType == "auth_error" {
				return m, m.setError("login incorrect")
			}
			return m, m.setError(fmt.Sprintf("error: %s", msg.resp.Description))

		case "auth_message":
			switch msg.resp.AuthMessageType {
			case "secret", "visible":
				answer := m.password.Value()
				req := &greetd.Request{
					Type:     "post_auth_message_response",
					Response: &answer,
				}
				if err := req.Encode(m.conn); err != nil {
					m.cancelAndClose()
					m.starting = false
					m.password.SetValue("")
					return m, m.setError(fmt.Sprintf("failed to respond: %s", err))
				}
				m.waitingForPAM = true
				return m, m.waitForResponse()

			case "info", "error":
				req := &greetd.Request{
					Type:     "post_auth_message_response",
					Response: nil,
				}
				if err := req.Encode(m.conn); err != nil {
					m.cancelAndClose()
					m.starting = false
					m.password.SetValue("")
					return m, m.setError(fmt.Sprintf("failed to respond: %s", err))
				}
				m.waitingForPAM = true
				var cmd tea.Cmd
				if msg.resp.AuthMessageType == "error" {
					cmd = m.setError(msg.resp.AuthMessage)
				}
				return m, tea.Batch(m.waitForResponse(), cmd)

			default:
				m.cancelAndClose()
				m.starting = false
				m.password.SetValue("")
				cmd := m.setError(fmt.Sprintf("unknown auth message type: %s", msg.resp.AuthMessageType))
				return m, cmd
			}

		default:
			m.cancelAndClose()
			m.starting = false
			m.password.SetValue("")
			cmd := m.setError(fmt.Sprintf("unexpected response: %s", msg.resp.Type))
			return m, cmd
		}

	case clearErrorMsg:
		m.errorMsg = ""
		return m, nil
	}

	var cmd tea.Cmd
	switch m.focused {
	case focusUsername:
		m.username, cmd = m.username.Update(msg)
	case focusPassword:
		m.password, cmd = m.password.Update(msg)
	case focusCommand:
		m.command, cmd = m.command.Update(msg)
	}
	return m, cmd
}

func (m *Model) updateFocus() {
	m.username.Blur()
	m.password.Blur()
	m.command.Blur()

	switch m.focused {
	case focusUsername:
		m.username.Focus()
	case focusPassword:
		m.password.Focus()
	case focusCommand:
		m.command.Focus()
	}
}

func (m *Model) waitForResponse() tea.Cmd {
	return func() tea.Msg {
		resp, err := greetd.DecodeResponse(m.conn)
		return greetdResponseMsg{resp: resp, err: err}
	}
}

func (m *Model) cancelAndClose() {
	if m.conn != nil {
		req := &greetd.Request{Type: "cancel_session"}
		req.Encode(m.conn)
		m.conn.Close()
		m.conn = nil
	}
}

func (m *Model) setError(msg string) tea.Cmd {
	m.errorMsg = msg
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}
