package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/xhos/yawn/internal/greetd"
)

type server struct {
	path     string
	listener net.Listener
	username string
	password string
}

type session struct {
	username      string
	authenticated bool
}

func newServer(path, username, password string) (*server, error) {
	if path == "" {
		path = filepath.Join(os.TempDir(), fmt.Sprintf("greetd-stub-%d.sock", os.Getpid()))
	}

	os.Remove(path)

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	return &server{
		path:     path,
		listener: listener,
		username: username,
		password: password,
	}, nil
}

func (s *server) close() {
	s.listener.Close()
	os.Remove(s.path)
}

func (s *server) serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *server) handle(conn net.Conn) {
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

func (s *server) process(req *greetd.Request, state *session) *greetd.Response {
	switch req.Type {
	case "create_session":
		state.username = req.Username
		return &greetd.Response{
			Type:            "auth_message",
			AuthMessageType: "secret",
			AuthMessage:     "Password: ",
		}

	case "post_auth_message_response":

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
