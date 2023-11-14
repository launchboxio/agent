package mux_socket

import "encoding/json"

type ActionCableEvent struct {
	Identifier map[string]string
	Data       *interface{}
	Command    string
	Type       string
	Message    json.RawMessage
}

type RawMessage struct {
	Type       string          `json:"type,omitempty"`
	Message    json.RawMessage `json:"message,omitempty"`
	Identifier string          `json:"identifier,omitempty"`
	Command    string          `json:"command,omitempty"`
}
