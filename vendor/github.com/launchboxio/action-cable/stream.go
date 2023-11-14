package mux_socket

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"time"
)

var ignoredEvents = []string{
	"ping",
	"welcome",
	"confirm_subscription",
}

type Stream struct {
	Url    string
	Header http.Header

	NotFoundHandler HandlerFunc
	ErrorHandler    func(error)
	OnDisconnect    func()
	OnConnect       func()
	OnMessage       func(message []byte)

	adapter       ActionCableAdapter
	subscriptions []*Subscription

	send chan []byte

	connected  bool
	connection *websocket.Conn
}

type Handler struct {
	HandlerFunc HandlerFunc
	Format      interface{}
}

func NewHandler(f HandlerFunc, format interface{}) *Handler {
	return &Handler{
		HandlerFunc: f,
		Format:      format,
	}
}

type HandlerFunc func(event *ActionCableEvent)

type Route struct {
	Channel string
	Handler HandlerFunc
}

func New(url string, header http.Header) (*Stream, error) {
	return &Stream{
		Url:           url,
		Header:        header,
		subscriptions: []*Subscription{},
		adapter:       ActionCableAdapter{},
	}, nil
}

// Connect connects to the websocket stream, and
// dispatches any received messages
func (s *Stream) Connect(ctx context.Context) error {
	u, err := url.Parse(s.Url)
	if err != nil {
		return err
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), s.Header)
	if err != nil {
		return err
	}
	done := make(chan struct{})

	s.onConnect(c)
	s.send = make(chan []byte, 10)

	defer c.Close()

	go func() {
		defer close(done)
		s.listen()
	}()

	for _, subscription := range s.subscriptions {
		if err := subscription.Connect(); err != nil {
			if s.ErrorHandler != nil {
				s.ErrorHandler(err)
			}
		}
	}

	for {
		select {
		case <-done:
			return nil
		case msg := <-s.send:
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				if s.ErrorHandler != nil {
					s.ErrorHandler(err)
				}
			}
		case <-ctx.Done():
			return s.close(done)
		}
	}
}

func (s *Stream) onConnect(connection *websocket.Conn) {
	s.connection = connection
	s.connected = true

	// Notify anyOnConnect handlers
	if s.OnConnect != nil {
		s.OnConnect()
	}
}

func (s *Stream) close(done chan struct{}) error {
	err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	s.OnDisconnect()
	select {
	case <-done:
	case <-time.After(time.Second):
	}
	return nil
}

func (s *Stream) listen() {
	for {
		_, message, err := s.connection.ReadMessage()
		if err != nil {
			s.ErrorHandler(err)
			continue
		}
		if s.OnMessage != nil {
			s.OnMessage(message)
		}
		if err = s.process(message); err != nil && s.ErrorHandler != nil {
			s.ErrorHandler(err)
		}
		continue
	}
}

// process parses an incoming message using the configured
// adapter, and then passes it along for dispatching
func (s *Stream) process(message []byte) error {
	// Parse the message using our adapter, t
	event, err := s.adapter.UnmarshalEvent(message)
	if err != nil {
		return err
	}

	return s.dispatch(event)
}

// dispatch takes a parsed event, and passes it off
// to all of the registered handlers for its channel
func (s *Stream) dispatch(event *ActionCableEvent) error {
	// TODO: Rather than route on event.Type, we should instead
	// get the channel Identifier
	// Ignore specified events to cut down chatter

	if slices.Contains(ignoredEvents, event.Type) {
		return nil
	}
	for _, subscription := range s.subscriptions {
		if reflect.DeepEqual(subscription.Identifier, event.Identifier) {
			subscription.Dispatch(event)
		}
	}
	return nil
}

func (s *Stream) IsConnected() bool {
	return s.connected
}

func (s *Stream) Send(event *ActionCableEvent) error {
	bytes, err := s.adapter.MarshalEvent(event)
	if err != nil {
		s.ErrorHandler(err)
		return err
	}

	s.send <- bytes
	return nil
}

func (s *Stream) SendRaw(bytes []byte) {
	s.send <- bytes
}

func (s *Stream) Subscribe(sub *Subscription) {
	sub.Stream = s
	s.subscriptions = append(s.subscriptions, sub)
}
