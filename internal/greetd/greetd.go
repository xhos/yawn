// Package greetd implements the greetd IPC protocol codec.
// spec: https://man.archlinux.org/man/greetd-ipc.7.en
package greetd

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

type Request struct {
	Type     string   `json:"type"`
	Username string   `json:"username,omitempty"`
	Response *string  `json:"response,omitempty"`
	Cmd      []string `json:"cmd,omitempty"`
	Env      []string `json:"env,omitempty"`
}

type Response struct {
	Type            string `json:"type"`
	ErrorType       string `json:"error_type,omitempty"`
	Description     string `json:"description,omitempty"`
	AuthMessageType string `json:"auth_message_type,omitempty"`
	AuthMessage     string `json:"auth_message,omitempty"`
}

func (r *Request) Encode(w io.Writer) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (r *Response) Encode(w io.Writer) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func DecodeRequest(r io.Reader) (*Request, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func DecodeResponse(r io.Reader) (*Response, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
