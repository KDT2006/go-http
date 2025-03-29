package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/KDT2006/go-http/internal/headers"
	"github.com/KDT2006/go-http/internal/request"
	"github.com/KDT2006/go-http/internal/response"
	"github.com/KDT2006/go-http/internal/server"
)

const port = 42069

func main() {
	// Custom Handler func
	customHandlerFunc := func(w *response.Writer, req *request.Request) *server.HandleError {
		target := req.RequestLine.RequestTarget
		switch {
		case target == "/yourproblem":
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
		case target == "/myproblem":
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

		// handle proxy
		case strings.HasPrefix(target, "/httpbin/"):
			fmt.Println(target)

			target := strings.TrimPrefix(target, "/httpbin/")

			resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", target))
			if err != nil {
				log.Println("error: http.Get() failed for proxying:", err)
				return &server.HandleError{
					StatusCode: response.InternalServerErrror,
					Message:    err.Error(),
				}
			}
			defer resp.Body.Close()

			// Remove Content-Length and add Transfer-Encoding header
			resp.Header.Del("Content-Length")
			resp.Header.Add("Transfer-Encoding", "chunked")
			// Announce X-Content-SHA256 and X-Content-Length as trailers in the Trailer header
			resp.Header.Add("Trailer", "X-Content-SHA256")
			resp.Header.Add("Trailer", "X-Content-Length")

			// Buffer for storing all of body
			bodyBuf := bytes.Buffer{}

			// Read chunks from response
			for {
				buf := make([]byte, 1024)
				n, err := resp.Body.Read(buf)
				if err != nil {
					if err == io.EOF {
						log.Println("Successfully read and transferred all chunks to client")

						// Calculate hash of the full response body and add the Trailers
						bodyHash := sha256.Sum256(bodyBuf.Bytes())

						trailers := headers.NewHeaders()
						trailers["X-Content-SHA256"] = fmt.Sprintf("%x", bodyHash)
						trailers["X-Content-Length"] = fmt.Sprintf("%d", bodyBuf.Len())

						err := w.WriteTrailers(trailers)
						if err != nil {
							log.Println("error: w.WriteTrailers() failed:", err)
						}

						return nil
					}

					log.Println("error: resp.Body.Read() failed:", err)
					return &server.HandleError{
						StatusCode: response.InternalServerErrror,
						Message:    err.Error(),
					}
				}
				fmt.Printf("Read %d bytes from resp\n", n)

				// Write response back to client
				_, err = w.Conn.Write(buf[:n])
				if err != nil {
					log.Println("error: conn.Write() failed writing resp back to client:", err)
					return nil
				}

				// Append to bodyBuf
				_, err = bodyBuf.Write(buf)
				if err != nil {
					log.Println("error: bodyBuf.Write() failed appending to buffer:", err)
					return &server.HandleError{
						StatusCode: response.InternalServerErrror,
						Message:    err.Error(),
					}
				}
			}

		// handle video
		case strings.HasPrefix(target, "/video"):
			// Check if the request method is GET
			if req.RequestLine.Method == "GET" {
				data, err := os.ReadFile("assets/vim.mp4")
				if err != nil {
					log.Println("error: os.ReadFile() failed when reading the video into memory:", err)
					return &server.HandleError{
						StatusCode: response.InternalServerErrror,
						Message:    err.Error(),
					}
				}

				// Populate the response with necessary data
				w.Headers = response.GetDefaultHeaders(len(data))
				w.Headers.Replace("Content-Type", "video/mp4")
				w.Status = response.OK
				w.Body = data

				// Write the Status line
				err = w.WriteStatusLine()
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

		return nil
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
