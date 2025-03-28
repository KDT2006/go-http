package response

import (
	"fmt"
	"io"
	"log"

	"example.com/httpfromtcp/internal/headers"
)

type StatusCode int
type WriterState int

const (
	OK = iota
	BadRequest
	InternalServerErrror
)

const (
	StatusLine = iota
	Headers
	Body
)

type Writer struct {
	Conn        io.Writer
	Headers     headers.Headers
	Status      StatusCode
	Body        []byte
	WriterState WriterState
}

func (w *Writer) WriteStatusLine() error {
	// Check for proper response order
	if w.WriterState != StatusLine {
		return fmt.Errorf("error: Improper response order, expected: Status Line -> Headers -> Body\n")
	}

	switch w.Status {
	case OK:
		_, err := w.Conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}

	case BadRequest:
		_, err := w.Conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}

	case InternalServerErrror:
		_, err := w.Conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}

	default:
		return nil
	}

	w.WriterState = Headers

	return nil
}

func (w *Writer) WriteHeaders() error {
	// Check for proper response order
	if w.WriterState != Headers {
		return fmt.Errorf("error: Improper response order, expected: Status Line -> Headers -> Body\n")
	}

	for key, value := range w.Headers {
		_, err := w.Conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	w.Conn.Write([]byte("\r\n")) // Final CRLF to denote end of headers

	w.WriterState = Body

	return nil
}

func (w *Writer) WriteBody() (int, error) {
	// Check for proper response order
	if w.WriterState != Body {
		return 0, fmt.Errorf("error: Improper response order, expected: Status Line -> Headers -> Body\n")
	}

	_, err := w.Conn.Write(w.Body)
	if err != nil {
		log.Println("error: WriteBody() failed:", err)
		return 0, err
	}

	return len(w.Body), nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = fmt.Sprint(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}
