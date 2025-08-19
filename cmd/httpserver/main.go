package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/JuanMartinCoder/http_protocol/internal/headers"
	"github.com/JuanMartinCoder/http_protocol/internal/request"
	"github.com/JuanMartinCoder/http_protocol/internal/response"
	"github.com/JuanMartinCoder/http_protocol/internal/server"
)

const port = 42069

func respond400() []byte {
	return []byte(`
				<html>
	  				<head>
    					<title>400 Bad Request</title>
  					</head>
  					<body>
   						<h1>Bad Request</h1>
    					<p>Your request honestly kinda sucked.</p>
  					</body>
				</html>`)
}

func respond500() []byte {
	return []byte(`
				<html>
	  				<head>
    					<title>500 Internal Server Error</title>
  					</head>
  					<body>
   						<h1>Internal Server Error</h1>
    					<p>Okay, you know what? This one is on me.</p>
  					</body>
				</html>`)
}

func respond200() []byte {
	return []byte(`
				<html>
	  				<head>
    					<title>200 OK</title>
  					</head>
  					<body>
   						<h1>Success!</h1>
    					<p>Your request was an absolute banger.</p>
  					</body>
				</html>`)
}

func toStr(b []byte) string {
	out := ""
	for _, v := range b {
		out += fmt.Sprintf("%02x", v)
	}
	return out
}

func main() {
	server, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		if req.RequestLine.RequestTarget == "/yourproblem" {
			w.WriteStatusLine(response.StatusBadRequestError)
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			w.WriteStatusLine(response.StatusInternalServerError)
			body = respond500()
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			resp, err := http.Get("https://httpbin.org/" + req.RequestLine.RequestTarget[len("/httpbin/"):])
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				body = respond500()
			} else {

				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteStatusLine(response.StatusOk)
				w.WriteHeaders(h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := resp.Body.Read(data)
					if err != nil {
						break
					}
					fullBody = append(fullBody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				trailers := headers.NewHeaders()

				out := sha256.Sum256(fullBody)

				trailers.Set("X-Content-SHA256", toStr(out[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailers)
				return
			}
		} else if req.RequestLine.RequestTarget == "/video" {
			w.WriteStatusLine(response.StatusOk)
			h.Replace("Content-Type", "video/mp4")
			file, err := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(file)))
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				w.WriteBody(respond500())
			}
			w.WriteHeaders(h)
			w.WriteBody(file)

		}
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")

		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(h)
		w.WriteBody(body)
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
