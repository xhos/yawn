package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xhos/yawn/internal/tui"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("yawn could not yawn: %v", err)
	}
}

func run() error {
	var cmd string
	var user string
	var width int
	var preauth bool
	var color string
	var minimal bool

	flag.StringVar(&cmd, "cmd", "", "command to run")
	flag.StringVar(&user, "user", "", "hardcodes a username to auth as")
	flag.IntVar(&width, "width", 8, "width of the input fields")
	flag.BoolVar(&preauth, "preauth", false, "start the auth loop immediately")
	flag.StringVar(&color, "color", "f5f5f5", "accent color")
	flag.BoolVar(&minimal, "minimal", false, "hide help and command display")

	flag.Parse()

	sock := os.Getenv("GREETD_SOCK")
	if sock == "" {
		return fmt.Errorf("GREETD_SOCK not found")
	}

	m := tui.InitialModel(sock, cmd, user, width, preauth, color, minimal)

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return err
	}

	return nil
}
