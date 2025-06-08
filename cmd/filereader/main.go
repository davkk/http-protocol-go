package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(file io.ReadCloser) <-chan string {
	messages := make(chan string)
	go func() {
		defer close(messages)
		line := ""
		for {
			data := make([]byte, 8, 8)

			n, err := file.Read(data)
			if err != nil {
				if line != "" {
					messages <- line
					line = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				break
			}

			str := string(data[:n])
			parts := strings.Split(str, "\n")

			for idx := 0; idx < len(parts)-1; idx++ {
				messages <- (line + parts[idx])
				line = ""
			}
			line += parts[len(parts)-1]
		}
	}()

	return messages
}

func main() {
	file, _ := os.Open("messages.txt")
	defer file.Close()

	for msg := range getLinesChannel(file) {
		fmt.Printf("read: %s\n", msg)
	}
}
