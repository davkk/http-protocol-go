package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

// Uppercase letters: A-Z
// Lowercase letters: a-z
// Digits: 0-9
// Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~

const validHeaderChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-.^_`|~"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	eol := bytes.Index(data, []byte(crlf))

	if eol < 0 {
		return 0, false, nil
	}

	if eol == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:eol], []byte(":"), 2)

	key := string(parts[0])
	if key != strings.TrimRight(key, " ") {
		return 0, false, errors.New("invalid header key")
	}

	key = strings.TrimSpace(key)
	value := string(bytes.TrimSpace(parts[1]))

	key = strings.ToLower(key)
	for _, c := range key {
		if !strings.ContainsRune(validHeaderChars, c) {
			return 0, false, errors.New("invalid character in header key")
		}
	}

	if header, exists := h[key]; exists {
		h[key] = fmt.Sprintf("%s, %s", header, value)
	} else {
		h[key] = value
	}
	return eol + 2, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	value, ok := h[strings.ToLower(key)]
	return value, ok
}

func (h Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	h[key] = value
}
