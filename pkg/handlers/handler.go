package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/centrifugal/centrifuge-go"
	"k8s.io/client-go/kubernetes"
	"log"
)

type Handler struct {
	Client *kubernetes.Clientset
}

func New(client *kubernetes.Clientset) *Handler {
	return &Handler{Client: client}
}

type GenericEvent struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

func (h *Handler) ProcessEvent(e centrifuge.PublicationEvent) error {
	event := &GenericEvent{}
	err := json.Unmarshal(e.Data, event)
	if err != nil {
		log.Println(err)
		return err
	}
	switch event.Event {
	case "apply_manifest":
		return h.applyManifest(event.Payload)
	case "delete_manifest":
		return h.deleteManifest(event.Payload)
	case "query":
		return h.query(event.Payload)
	default:
		return errors.New(fmt.Sprintf("Unsupported event '%s'", event.Event))
	}
}
