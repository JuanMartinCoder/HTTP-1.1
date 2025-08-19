package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/JuanMartinCoder/http_protocol/internal/request"
	"github.com/JuanMartinCoder/http_protocol/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

type (
	Handler func(w response.Writer, req *request.Request)
	Server  struct {
		listener net.Listener
		handler  Handler
		closed   atomic.Bool
	}
)

func Serve(port int, handlerFunc Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: l,
		handler:  handlerFunc,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error accepting conn: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequestError)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	s.handler(*responseWriter, r)
}
