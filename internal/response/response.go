package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"http-protocol-go/internal/headers"
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

type WriterState int

const (
	WriterStatusLine WriterState = iota
	WriterHeaders
	WriterBody
	WriterTrailers
)

type Writer struct {
	writer io.Writer
	state  WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  WriterStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != WriterStatusLine {
		return errors.New("wrong state to write status line")
	}
	defer func() { w.state = WriterHeaders }()

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

	_, err := w.writer.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != WriterHeaders {
		return errors.New("wrong state to write headers")
	}
	defer func() { w.state = WriterBody }()

	for key, value := range h {
		w.writer.Write(fmt.Appendf(nil, "%s: %s\r\n", key, value))
	}
	w.writer.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != WriterBody {
		return 0, errors.New("wrong state to write body")
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != WriterBody {
		return 0, errors.New("wrong state to write chunked body")
	}

	if _, err := fmt.Fprintf(w.writer, "%x\r\n", len(p)); err != nil {
		return 0, err
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}

	if _, err := w.writer.Write([]byte("\r\n")); err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != WriterBody {
		return 0, errors.New("wrong state to write chunked body done")
	}
	w.state = WriterTrailers
	return fmt.Fprintf(w.writer, "0\r\n")
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != WriterTrailers {
		return errors.New("wrong state to write trailers")
	}

	for key, value := range h {
		_, err := w.writer.Write(fmt.Appendf(nil, "%s: %s\r\n", key, value))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
