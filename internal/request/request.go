package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"example.com/httpfromtcp/internal/headers"
)

const (
	bufferSize = 8
)

const (
	INITIALIZED = iota
	DONE
	PARSING_HEADERS
	PARSING_BODY
)

type Request struct {
	RequestLine RequestLine
	State       int
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 0, bufferSize)
	request := &Request{
		State: INITIALIZED,
	}

	for request.State != DONE {
		// Temporary buffer for reading
		tmpBuf := make([]byte, bufferSize)
		n, err := reader.Read(tmpBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Append new data to accumulated buffer
		buf = append(buf, tmpBuf[:n]...)

		// Try to parse
		parsed, err := request.parse(buf)
		if err != nil {
			return nil, err
		}

		// fmt.Printf("%+v\n", *request)

		if parsed > 0 {
			buf = buf[parsed:]
		}
	}

	if request.State != DONE {
		return nil, fmt.Errorf("incomplete request")
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.State != DONE {
		// fmt.Println("Size of data: ", len(data))
		// fmt.Println("totalBytesParsed: ", totalBytesParsed)
		n, err := r.parseSingle(data[totalBytesParsed:])
		// fmt.Println("n: ", n)
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += n

		if n == 0 {
			break // need more data
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case INITIALIZED:
		n, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n > 0 {
			r.State = PARSING_HEADERS
		}

		return n, nil

	case PARSING_HEADERS:
		// Initialize headers if not already
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		// Parse headers
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		// Move to next state if headers are done
		if done {
			r.State = PARSING_BODY
		}

		// Return if no data was parsed
		return n, nil

	case PARSING_BODY:
		// Check for Content-Length header(which indicates a body)
		if r.Headers.Get("Content-Length") == "" {
			r.State = DONE
			return 0, nil
		}

		contentLengthInt, err := strconv.ParseInt(r.Headers.Get("Content-Length"), 10, 64)
		// log.Printf("\r\nContent-Length: %d\r\n", contentLengthInt)
		if err != nil {
			return 0, err
		}
		// fmt.Println("Content-Length: ", contentLengthInt)

		// Append all the data to the requests .Body field
		r.Body = append(r.Body, data...)

		// If the length of the body is greater than the Content-Length header, return an error
		if len(r.Body) > int(contentLengthInt) {
			return 0, fmt.Errorf("error: body length exceeds Content-Length header")
		}

		// If the length of the body is equal to the Content-Length header, move to the done state
		if len(r.Body) == int(contentLengthInt) {
			r.State = DONE
		}

		// Report that you've consumed the entire length of the data you were given
		return len(data), nil

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	index := bytes.Index(data, []byte("\r\n"))
	if index == -1 {
		return 0, nil // need more data
	}

	if index == 0 {
		return 0, nil // empty line
	}

	parts := strings.Split(string(data[:index]), " ")

	// Check correct split
	if len(parts) != 3 {
		return 0, fmt.Errorf("error malformed request line: %s", data)
	}

	// Validate method (all uppercase)
	for _, char := range parts[0] {
		if char < 'A' || char > 'Z' {
			return 0, fmt.Errorf("error unknown request method: %s", parts[0])
		}
	}

	// Check HTTP version
	if parts[2] != "HTTP/1.1" {
		return 0, fmt.Errorf("error unsupported HTTP version: %s", parts[2])
	}

	// Store parsed data
	r.RequestLine = RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   "1.1", // Explicitly set to "1.1" as in your test
	}

	return index + 2, nil // Account for \r\n
}
