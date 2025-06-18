package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http-protocol-go/internal/request"
	"http-protocol-go/internal/response"
	"http-protocol-go/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		if strings.HasPrefix(req.RequestLine.Target, "/httpbin") {
			proxy(w, req)
			return
		}
		switch req.RequestLine.Target {
		case "/video":
			responseVideo(w)
		case "/yourproblem":
			response400(w)
		case "/myproblem":
			response500(w)
		default:
			response200(w)
		}
	})

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

func response400(w *response.Writer) {
	w.WriteStatusLine(response.BAD_REQUEST)

	body := `<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>`

	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}

func response500(w *response.Writer) {
	w.WriteStatusLine(response.INTERNAL_SERVER_ERROR)
	body := `<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
	</body>
</html>`

	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}

func response200(w *response.Writer) {
	w.WriteStatusLine(response.OK)

	body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}

func proxy(w *response.Writer, req *request.Request) {
	route := strings.TrimPrefix(req.RequestLine.Target, "/httpbin")

	url := fmt.Sprintf("http://httpbin.org%s", route)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error while proxying request: %v", err)
		response500(w)
		return
	}
	defer resp.Body.Close()

	newHeaders := response.GetDefaultHeaders(0)
	for key, values := range resp.Header {
		for _, value := range values {
			newHeaders.Set(key, value)
		}
	}

	newHeaders.Del("Content-Length")
	newHeaders.Set("Transfer-Encoding", "chunked")

	w.WriteStatusLine(response.OK)
	w.WriteHeaders(newHeaders)

	body := make([]byte, 0)
	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			response500(w)
			break
		}

		if n == 0 {
			break
		}

		if _, err = w.WriteChunkedBody(buf[:n]); err != nil {
			response500(w)
			break
		}

		body = append(body, buf[:n]...)
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		response500(w)
		return
	}

	trailers := response.GetDefaultHeaders(0)
	trailers.Del("Content-Length")
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(body)))
	trailers.Set("X-Content-Length", fmt.Sprint(len(body)))

	err = w.WriteTrailers(trailers)
	if err != nil {
		response500(w)
		return
	}
}

func responseVideo(w *response.Writer) {
	w.WriteStatusLine(response.OK)
	data, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		response500(w)
		return
	}

	h := response.GetDefaultHeaders(len(data))
	h.Set("Content-Type", "video/mp4")
	w.WriteHeaders(h)

	w.WriteBody(data)
}
