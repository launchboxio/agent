package mux_socket

import (
	"encoding/json"
)

type ActionCableAdapter struct {
}

// TODO: We should handle disconnects...
func (aca *ActionCableAdapter) MarshalEvent(event *ActionCableEvent) ([]byte, error) {
	message := &RawMessage{
		Command: event.Command,
		Type:    event.Type,
	}
	if event.Identifier != nil {
		identBytes, err := json.Marshal(event.Identifier)
		if err != nil {
			return nil, err
		}
		message.Identifier = string(identBytes)
	}
	if event.Data != nil {
		dataBytes, err := json.Marshal(event.Data)
		if err != nil {
			return nil, err
		}
		message.Message = dataBytes
	}

	return json.Marshal(message)
}

func (aca *ActionCableAdapter) UnmarshalEvent(message []byte) (*ActionCableEvent, error) {
	rawMessage := &RawMessage{}
	err := json.Unmarshal(message, rawMessage)
	if err != nil {
		return nil, err
	}

	ace := &ActionCableEvent{
		Command: rawMessage.Command,
		Type:    rawMessage.Type,
		Message: rawMessage.Message,
	}
	if rawMessage.Identifier != "" {
		identifier := map[string]string{}
		err := json.Unmarshal([]byte(rawMessage.Identifier), &identifier)
		if err != nil {
			return nil, err
		}
		ace.Identifier = identifier
	}

	return ace, err
}
