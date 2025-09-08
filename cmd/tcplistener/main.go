package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

const BUFFER_SIZE = 8

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {

		con, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("New connection: ", con.RemoteAddr())

		req, err := request.RequestFromReader(con)
		if err != nil {
			log.Fatal(err)
		}

		rl := req.RequestLine
		h := req.Headers
		b := req.Body
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", rl.Method)
		fmt.Printf("- Target: %s\n", rl.RequestTarget)
		fmt.Printf("- Version: %s\n", rl.HttpVersion)
		fmt.Printf("Headers:\n")
		for k, v := range h {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", b)

		con.Close()
		fmt.Println("Connection closed")
	}
}
