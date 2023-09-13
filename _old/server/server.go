package server

import "net/http"

const DefaultAddr = ":42069"

type Server struct {
	*http.Server

	Handler http.Handler
	Tasks map[]
}

func NewServer() *Server {
	s := &Server{}
	s.Server = &http.Server{
		Addr:    DefaultAddr,
		Handler: s,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
