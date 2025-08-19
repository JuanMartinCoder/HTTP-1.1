package main

import (
	"fmt"
	"log"
	"net"

	"github.com/JuanMartinCoder/http_protocol/internal/request"
)

func showRequest(req *request.Request) {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

	headers := req.Headers.GetAllHeaders()
	fmt.Println("Headers:")
	for key, val := range headers {
		fmt.Printf("- %s: %s\n", key, val)
	}
	fmt.Println("Body:")
	fmt.Println(req.Body)
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection has been accepted")

		go func(c net.Conn) {
			r, err := request.RequestFromReader(c)
			if err != nil {
				log.Fatal(err)
			}
			showRequest(r)
		}(conn)

	}
}
