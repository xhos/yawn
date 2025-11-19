package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/xhos/yawn/internal/greetd"
)

const username = "test"
const password = "test"

func main() {
	server, err := NewServer("", username, password)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}
	defer server.Close()

	fmt.Printf("socket:   %s\n", server.Path())
	fmt.Printf("user:     %s\npassword: %s\n", username, password)
	fmt.Printf("\nrun yawn with:\nGREETD_SOCK=%s go run ./cmd/yawn/main.go\n\n", server.Path())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Serve(); err != nil {
			log.Print(err)
		}
	}()

	<-sigChan
}

type Server struct {
	path     string
	listener net.Listener
	username string
	password string
}

type session struct {
	username      string
	authenticated bool
}

func NewServer(path, username, password string) (*Server, error) {
	if path == "" {
		path = filepath.Join(os.TempDir(), "greetd-stub.sock")
	}

	os.Remove(path)

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	return &Server{
		path:     path,
		listener: listener,
		username: username,
		password: password,
	}, nil
}

func (s *Server) Path() string {
	return s.path
}

func (s *Server) Close() error {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.path)
	return nil
}

func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	state := &session{}

	for {
		req, err := greetd.DecodeRequest(conn)
		if err != nil {
			return
		}

		resp := s.process(req, state)
		if resp == nil {
			return
		}

		if err := resp.Encode(conn); err != nil {
			return
		}
	}
}

func (s *Server) process(req *greetd.Request, state *session) *greetd.Response {
	switch req.Type {
	case "create_session":
		state.username = req.Username
		state.authenticated = false
		return &greetd.Response{
			Type:            "auth_message",
			AuthMessageType: "secret",
			AuthMessage:     "Password: ",
		}

	case "post_auth_message_response":
		if state.username == "" {
			return &greetd.Response{
				Type:        "error",
				ErrorType:   "error",
				Description: "no session",
			}
		}

		if state.authenticated {
			return &greetd.Response{
				Type:        "error",
				ErrorType:   "error",
				Description: "already authenticated",
			}
		}

		if state.username != s.username || req.Response == nil || *req.Response != s.password {
			state.username = ""
			return &greetd.Response{
				Type:        "error",
				ErrorType:   "auth_error",
				Description: "authentication failed",
			}
		}

		state.authenticated = true
		return &greetd.Response{Type: "success"}

	case "start_session":
		if !state.authenticated {
			return &greetd.Response{
				Type:        "error",
				ErrorType:   "error",
				Description: "not authenticated",
			}
		}

		log.Printf("starting session: %v", req.Cmd)
		return &greetd.Response{Type: "success"}

	case "cancel_session":
		return nil

	default:
		return &greetd.Response{
			Type:        "error",
			ErrorType:   "error",
			Description: fmt.Sprintf("unknown request: %s", req.Type),
		}
	}
}
