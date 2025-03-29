# HTTP Parser & Response Sender in Go
  
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A **zero-dependency**, **stdlib-only** HTTP/1.1 request parser and response sender written from scratch in Go. Built as a learning project, this library provides low-level control over HTTP message handling while supporting features like:

- **Full HTTP/1.1 Request & Response Flow**  
- **Chunked Transfer Encoding**  
- **Trailer Support**  
- **Binary Data Handling**

## **Why?**

Most Go HTTP libraries abstract away low-level details. This project was built to:

- Learn the internals of HTTP/1.1
- Experiment with raw TCP-based HTTP handling

## **Quick Start**

### Clone the repo:

```bash
git clone https://github.com/KDT2006/go-http.git
cd go-http
```

### Run the server:

```bash
go run ./cmd/httpserver/main.go
```

## **Sameple Endpoints(main.go)**

### 1. Success

```bash
curl http://localhost:8080
```
Returns a 200 OK HTML response.

### 2. Custom Error (400):

```bash
curl -v http://localhost:8080/yourproblem
```
Returns a 400 error with HTML explaining the request "kinda sucked".

### 3. Server Error (500):

```bash
curl -v http://localhost:8080/myproblem
```
Returns a 500 error with HTML admitting fault.

### 4. Proxy to httpbin:

```bash
curl -v http://localhost:8080/httpbin/html
```
Proxies to https://httpbin.org/ with chunked encoding and trailers.

### 5. Stream a video: 

```bash
curl http://localhost:8080/video --output video.mp4
```
Streams an MP4 file (requires assets/vim.mp4 to exist).