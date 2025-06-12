package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"http-protocol-go/internal/request"
	"http-protocol-go/internal/response"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

func (he *HandlerError) Error() string {
	return fmt.Sprintln(he.StatusCode, he.Message)
}

func (he *HandlerError) Write(w *response.Writer) {
	w.WriteStatusLine(response.BAD_REQUEST)
	w.WriteHeaders(response.GetDefaultHeaders(len(he.Message)))
	w.WriteBody([]byte(he.Message))
}

type Handler func(w *response.Writer, req *request.Request)

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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)

	req, reqErr := request.RequestFromReader(conn)
	if reqErr != nil {
		err := &HandlerError{
			StatusCode: int(response.BAD_REQUEST),
			Message:    reqErr.Error(),
		}
		err.Write(w)
		return
	}

	// body := bytes.NewBuffer([]byte{})
	body := new(bytes.Buffer)

	s.handler(w, req)

	w.WriteStatusLine(response.OK)
	w.WriteHeaders(response.GetDefaultHeaders(len(body.Bytes())))
	w.WriteBody(body.Bytes())
}
