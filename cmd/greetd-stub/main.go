package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()

	server, err := newServer("", "test", "test")
	if err != nil {
		log.Fatalf("create server: %v", err)
	}
	defer server.close()

	fmt.Printf("socket:   %s\n", server.path)
	fmt.Printf("user:     test\npassword: test\n\n")

	go server.serve()

	cmd := exec.Command("go", append([]string{"run", "./cmd/yawn"}, flag.Args()...)...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GREETD_SOCK=%s", server.path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("yawn failed: %v", err)
	}
}
