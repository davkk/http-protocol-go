package response

import (
	"fmt"
	"http-protocol-go/internal/headers"
	"io"
	"strconv"
	"strings"
)

type StatusCode int

func (code StatusCode) Code() string {
	return strconv.Itoa(int(code))
}

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	segments := []string{"HTTP/1.1"}

	switch statusCode {
	case OK:
		segments = append(segments, OK.Code(), "OK")
	case BAD_REQUEST:
		segments = append(segments, BAD_REQUEST.Code(), "Bad Response")
	case INTERNAL_SERVER_ERROR:
		segments = append(segments, INTERNAL_SERVER_ERROR.Code(), "Internal Server Error")
	default:
		segments = append(segments, statusCode.Code())
	}

	statusLine := strings.Join(segments, " ") + "\r\n"

	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		w.Write(fmt.Appendf(nil, "%s: %s\r\n", key, value))
	}
	w.Write([]byte("\r\n"))
	return nil
}
