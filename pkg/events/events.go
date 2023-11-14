package events

const (
	ProjectCreatedEvent string = "projects.created"
	ProjectPausedEvent         = "projects.paused"
	ProjectResumedEvent        = "projects.resumed"
	ProjectUpdatedEvent        = "projects.updated"
	ProjectDeletedEvent        = "projects.deleted"
	AddonCreatedEvent          = "addons.created"
	AddonUpdatedEvent          = "addons.updated"
	AddonDeletedEvent          = "addons.deleted"
)

type LaunchboxEvent struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}
