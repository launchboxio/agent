package agent

type EventHandler struct {
}

type Event struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

func (e *EventHandler) Process(ev Event) {

}
