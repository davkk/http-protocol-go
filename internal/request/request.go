package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	"http-protocol-go/internal/headers"
)

const bufferSize = 8
const crlf = "\r\n"

const (
	INIT int = iota
	DONE
	READING_HEADERS
	READING_BODY
)

type Request struct {
	State       int
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion string
	Target      string
	Method      string
}

var METHODS = []string{"GET", "POST"}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	eol := bytes.Index(data, []byte(crlf))
	if eol < 0 {
		return nil, 0, nil
	}

	lines := strings.Split(string(data[:eol]), crlf)
	if len(lines) == 0 {
		return nil, 0, errors.New("data is empty")
	}

	request := lines[0]
	parts := strings.Split(request, " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("request line does not have 3 parts")
	}

	method := parts[0]
	if !slices.Contains(METHODS, method) {
		return nil, 0, errors.New("method not supported")
	}

	target := parts[1]

	version := parts[2]
	version = version[len(version)-3:]
	if version != "1.1" {
		return nil, 0, errors.New("http version must be 1.1")
	}

	return &RequestLine{
		Method:      method,
		Target:      target,
		HttpVersion: version,
	}, eol + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalParsed := 0

	for r.State != DONE {
		n, err := r.parseSingle(data[totalParsed:])
		if err != nil {
			return 0, err
		}

		totalParsed += n
		if n == 0 {
			break
		}
	}

	return totalParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case INIT:
		request, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *request
		r.State = READING_HEADERS
		return n, nil

	case READING_HEADERS:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			r.State = DONE
			return 0, err
		}
		if done {
			if _, exists := r.Headers.Get("Content-Length"); exists {
				r.State = DONE
			}
			r.State = READING_BODY
		}
		return n, nil

	case READING_BODY:
		contentLenHeader, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.State = DONE
			return len(data), nil
		}

		contentLen, err := strconv.Atoi(contentLenHeader)
		if err != nil {
			return 0, errors.New("invalid content length value")
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) > contentLen {
			return 0, errors.New("body length greater than content length")
		}

		if len(r.Body) == contentLen {
			r.State = DONE
		}

		return len(data), nil

	case DONE:
		return 0, errors.New("trying to read data in DONE state")

	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)

	readIdx := 0
	request := &Request{
		State:   INIT,
		Headers: headers.NewHeaders(),
	}

	for request.State != DONE {
		if readIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.State = DONE
				break
			}
			return nil, err
		}

		readIdx += bytesRead

		bytesParsed, err := request.parse(buf[:readIdx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[bytesParsed:])
		readIdx -= bytesParsed
	}

	return request, nil
}
