package events

import (
	"encoding/json"
	"github.com/go-logr/logr"
	action_cable "github.com/launchboxio/action-cable"
	launchbox "github.com/launchboxio/launchbox-go-sdk/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	Logger logr.Logger
	Client client.Client
	Sdk    *launchbox.Config
}

func New(logger logr.Logger, client client.Client, sdk *launchbox.Config) *Handler {
	handler := &Handler{
		Logger: logger,
		Client: client,
		Sdk:    sdk,
	}
	return handler
}

type HandlerFunc func(event *LaunchboxEvent) error

func (h *Handler) RegisterSubscriptions(stream *action_cable.Stream, identifier map[string]string) {
	subscription := action_cable.NewSubscription(identifier)

	projectHandler := h.projectHandler()
	addonHandler := h.addonHandler()

	subscription.Handler(func(event *action_cable.ActionCableEvent) {
		var handler HandlerFunc
		parsedEvent := &LaunchboxEvent{}
		err := json.Unmarshal(event.Message, parsedEvent)
		if err != nil {
			h.Logger.Error(err, "Failed parsing event")
		}
		switch parsedEvent.Type {
		case ProjectCreatedEvent:
			handler = projectHandler.Create
		case ProjectUpdatedEvent:
			handler = projectHandler.Update
		case ProjectDeletedEvent:
			handler = projectHandler.Delete
		case ProjectPausedEvent:
			handler = projectHandler.Pause
		case ProjectResumedEvent:
			handler = projectHandler.Resume
		case AddonCreatedEvent:
			handler = addonHandler.Create
		case AddonUpdatedEvent:
			handler = addonHandler.Update
		case AddonDeletedEvent:
			handler = addonHandler.Delete
		}
		if err := handler(parsedEvent); err != nil {
			h.Logger.Error(err, "Handler execution failed", "event", parsedEvent.Type)
		}
	})

	stream.Subscribe(subscription)
}

func (h *Handler) projectHandler() *ProjectHandler {
	return &ProjectHandler{
		Logger: h.Logger,
		Client: h.Client,
		Sdk:    h.Sdk,
	}
}

func (h *Handler) addonHandler() *AddonHandler {
	return &AddonHandler{
		Logger: h.Logger,
		Client: h.Client,
	}
}
