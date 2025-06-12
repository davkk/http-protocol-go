package main

import (
	"fmt"
	"net"
	"os"

	"http-protocol-go/internal/request"
	"http-protocol-go/pkg/assert"
)

func main() {
	list, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer list.Close()

	for {
		conn, err := list.Accept()
		assert.NoError(err, "connection error")

		// fmt.Println("[+] connection accepted")

		result, err := request.RequestFromReader(conn)
		assert.NoError(err, "parsing request failed")

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", result.RequestLine.Method)
		fmt.Printf("- Target: %s\n", result.RequestLine.Target)
		fmt.Printf("- Version: %s\n", result.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range result.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(result.Body))

		conn.Close()
		// fmt.Println("[-] connection closed")
	}
}
