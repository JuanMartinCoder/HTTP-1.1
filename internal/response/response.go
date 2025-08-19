package response

import (
	"fmt"
	"io"

	"github.com/JuanMartinCoder/http_protocol/internal/headers"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequestError     StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusOk:
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case StatusBadRequestError:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}

	case StatusInternalServerError:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}

	default:
		_, err := w.writer.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	allHeaders := headers.GetAllHeaders()

	for k, v := range allHeaders {
		headerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.writer.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return *headers
}
