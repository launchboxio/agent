package server

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	BindAddress string
}

func New(bindAddress string) *Server {
	return &Server{
		BindAddress: bindAddress,
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/healthz", health)
	http.HandleFunc("/readyz", ready)

	return http.ListenAndServe(s.BindAddress, nil)
}

func health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
func ready(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "OK")
}
