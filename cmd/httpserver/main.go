package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"http-protocol-go/internal/request"
	"http-protocol-go/internal/response"
	"http-protocol-go/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		switch req.RequestLine.Target {
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
