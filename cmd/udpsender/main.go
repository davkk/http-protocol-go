package main

import (
	"bufio"
	"fmt"
	"http-protocol/pkg/assert"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	assert.NoError(err, "udp resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(err, "udp dial failed")
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		conn.Write([]byte(str))
	}
}
