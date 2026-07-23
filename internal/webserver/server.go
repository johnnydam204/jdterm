package webserver

import (
	"fmt"
	"net/http"
)

type Server struct {
	addr string
}

func New(addr string) *Server {

	return &Server{
		addr: addr,
	}
}

func (s *Server) Start() error {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web"))

	mux.Handle("/", fs)

	fmt.Printf("JDTerm HTTP Server started\n")
	fmt.Printf("Open: http://localhost%s\n", s.addr)

	return http.ListenAndServe(s.addr, mux)

}
