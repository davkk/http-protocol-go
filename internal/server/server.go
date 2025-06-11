package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"http-protocol/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errors.New("failed to create server: " + err.Error())
	}

	server := &Server{
		listener: list,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() {
	s.closed.Store(true)
	s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			log.Printf("Error while accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

// HTTP/1.1 200 OK
// Content-Type: text/plain
//
// Hello World!
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
}
