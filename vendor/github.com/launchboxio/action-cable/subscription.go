package mux_socket

type Subscription struct {
	Identifier map[string]string
	Stream     *Stream

	handlers []HandlerFunc
}

func NewSubscription(identifier map[string]string) *Subscription {
	return &Subscription{Identifier: identifier}
}

func (s *Subscription) Dispatch(event *ActionCableEvent) error {
	for _, handler := range s.handlers {
		handler(event)
	}
	return nil
}

func (s *Subscription) Handler(handler HandlerFunc) {
	s.handlers = append(s.handlers, handler)
}

func (s *Subscription) Connect() error {
	event := &ActionCableEvent{
		Command:    "subscribe",
		Identifier: s.Identifier,
	}

	return s.Stream.Send(event)
}

func (s *Subscription) Send(event *ActionCableEvent) error {
	return s.Stream.Send(event)
}
