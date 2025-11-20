package tui

import (
	"net"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xhos/yawn/internal/greetd"
)

type greetdResponseMsg struct {
	resp *greetd.Response
	err  error
}

type clearErrorMsg struct{}

type focus int

const (
	focusUsername focus = iota
	focusPassword
	focusCommand
)

type Model struct {
	sockPath string
	conn     net.Conn

	username textinput.Model
	password textinput.Model
	command  textinput.Model

	focused        focus
	editingCommand bool
	waitingForPAM  bool
	starting       bool

	errorMsg string
	w, h     int
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func InitialModel(sockPath, cmd, username string, inputWidth int) *Model {
	usernameInput := textinput.New()
	usernameInput.Placeholder = "username"
	usernameInput.Width = inputWidth
	usernameInput.Prompt = ""
	usernameInput.PromptStyle = accentStyle
	usernameInput.TextStyle = accentStyle
	if username != "" {
		usernameInput.SetValue(username)
	}

	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.Width = inputWidth
	passwordInput.Prompt = ""
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'

	commandInput := textinput.New()
	commandInput.Placeholder = "session command"
	commandInput.Prompt = ""
	if cmd != "" {
		commandInput.SetValue(cmd)
	}

	m := &Model{
		sockPath: sockPath,
		username: usernameInput,
		password: passwordInput,
		command:  commandInput,
		focused:  focusUsername,
	}

	if username != "" {
		m.focused = focusPassword
	}

	m.updateFocus()
	return m
}
