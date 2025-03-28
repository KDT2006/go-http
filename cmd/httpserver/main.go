package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/httpfromtcp/internal/request"
	"example.com/httpfromtcp/internal/response"
	"example.com/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	// Custom Handler func
	customHandlerFunc := func(w *response.Writer, req *request.Request) *server.HandleError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			content := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
			w.Headers = response.GetDefaultHeaders(len(content))
			w.Headers.Replace("Content-Type", "text/html; charset=utf-8")
			w.Status = response.BadRequest
			w.Body = []byte(content)
			return &server.HandleError{
				StatusCode: response.BadRequest,
				Message:    content,
			}
		case "/myproblem":
			content := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
			w.Headers = response.GetDefaultHeaders(len(content))
			w.Headers.Replace("Content-Type", "text/html; charset=utf-8")
			w.Status = response.InternalServerErrror
			w.Body = []byte(content)
			return &server.HandleError{
				StatusCode: response.InternalServerErrror,
				Message:    content,
			}
		default:
			content := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
			w.Headers = response.GetDefaultHeaders(len(content))
			w.Headers.Replace("Content-Type", "text/html; charset=utf-8")
			fmt.Printf("%+v\n", w.Headers)
			w.Status = response.OK
			w.Body = []byte(content)

			// Write the Status line
			err := w.WriteStatusLine()
			if err != nil {
				log.Println("error: WriteStatusLine() failed:", err)
				return &server.HandleError{
					StatusCode: response.InternalServerErrror,
					Message:    err.Error(),
				}
			}

			// Create and write the response headers
			err = w.WriteHeaders()
			if err != nil {
				log.Println("error: response.WriteStatusLine() failed:", err)
				return &server.HandleError{
					StatusCode: response.InternalServerErrror,
					Message:    err.Error(),
				}
			}

			// Write the response body
			_, err = w.WriteBody()
			if err != nil {
				log.Println("error: conn.Write() failed:", err)
				return &server.HandleError{
					StatusCode: response.InternalServerErrror,
					Message:    err.Error(),
				}
			}
			return nil
		}
	}

	server, err := server.Serve(port, customHandlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
