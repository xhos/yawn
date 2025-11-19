package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

	flag.StringVar(&cmd, "cmd", "", "command to run")
	flag.StringVar(&user, "user", "", "hardcodes a username to auth as")
	flag.IntVar(&width, "width", 8, "width of the input fields")

	flag.Parse()

	sock := os.Getenv("GREETD_SOCK")
	if sock == "" {
		return fmt.Errorf("GREETD_SOCK not found")
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	defer conn.Close()

	m := tui.InitialModel(conn, cmd, user, width)

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return err
	}

	return nil
}
