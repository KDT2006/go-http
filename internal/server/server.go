package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"example.com/httpfromtcp/internal/request"
	"example.com/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Handler  HandlerFunc
	closed   atomic.Bool
}

type HandleError struct {
	StatusCode response.StatusCode
	Message    string
}

type HandlerFunc func(w *response.Writer, req *request.Request) *HandleError

func Serve(port int, handler HandlerFunc) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		Listener: ln,
		Handler:  handler,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.Listener.Close()
}

func (s *Server) listen() {
	// Accept new connections as long as the server's open
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			log.Println("Error accepting new connection:", err)
			continue
		}
		log.Println("New accpeted connection:", conn.RemoteAddr())

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse the request
	parsedReq, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("error: RequestFromReader() failed parsing the request:", err)
		s.writeError(&response.Writer{
			Conn:        conn,
			Status:      response.InternalServerErrror,
			Headers:     response.GetDefaultHeaders(0),
			WriterState: response.StatusLine,
		})
		return
	}

	// Buffer for the handler to write to
	buf := new(bytes.Buffer)

	// Call the handler and process the error if there's any
	responseWriter := &response.Writer{
		Conn:        buf,
		WriterState: response.StatusLine,
	}
	handlerErr := s.Handler(responseWriter, parsedReq)
	if handlerErr != nil {
		s.writeError(responseWriter)
	}

	// Write to connection regardless of error
	// as everything will be written to buf
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Println("error: conn.Write() failed:", err)
	}
}

func (s *Server) writeError(responseWriter *response.Writer) {
	// Write the HTTP status line
	err := responseWriter.WriteStatusLine()
	if err != nil {
		log.Println("Error writing to conn:", err)
		return
	}

	// Write the HTTP response headers
	err = responseWriter.WriteHeaders()
	if err != nil {
		log.Println("Error writing to conn:", err)
		return
	}

	// Write the HTTP body
	_, err = responseWriter.WriteBody()
	if err != nil {
		log.Println("Error writing to conn:", err)
		return
	}
}
