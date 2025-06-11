package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"http-protocol/internal/request"
	"http-protocol/internal/response"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

func (he *HandlerError) Error() string {
	return fmt.Sprintln(he.StatusCode, he.Message)
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errors.New("failed to create server: " + err.Error())
	}

	server := &Server{
		listener: list,
		handler:  handler,
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

	req, reqErr := request.RequestFromReader(conn)
	if reqErr != nil {
		return
	}

	body := new(bytes.Buffer)
	err := s.handler(body, req)
	if err != nil {
		response.WriteStatusLine(conn, response.BAD_REQUEST)
		response.WriteHeaders(conn, response.GetDefaultHeaders(len(err.Message)))
		conn.Write([]byte(err.Message))
		return
	}

	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(len(body.Bytes())))
	conn.Write(body.Bytes())
}
