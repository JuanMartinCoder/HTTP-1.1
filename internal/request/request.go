package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/JuanMartinCoder/http_protocol/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine    RequestLine
	Headers        *headers.Headers
	Body           string
	state          parserState
	bodyLenghtRead int
}

var (
	BAD_REQUEST_LINE_ERROR         = fmt.Errorf("malformed request-line")
	UNSUPPORTED_HTTP_VERSION_ERROR = fmt.Errorf("unsupported http version")
	CRLF                           = []byte("\r\n")
)

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	index := bytes.Index(b, CRLF)
	if index == -1 {
		return nil, 0, nil
	}

	startLine := b[:index]
	read := index + len(CRLF)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 { // check if the request Line is complete
		return nil, 0, BAD_REQUEST_LINE_ERROR
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" { // check if the request Line is complete
		return nil, 0, BAD_REQUEST_LINE_ERROR
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool {
	contentLength, ok := r.Headers.Get("content-length")
	if !ok {
		return false
	}
	contentLen, err := strconv.Atoi(contentLength)
	if err != nil {
		return false
	}
	return contentLen > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, fmt.Errorf("error in state")
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:

			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			contentLength, ok := r.Headers.Get("content-length")
			if !ok {
				r.state = StateDone
				return len(data), nil
			}
			contentLen, err := strconv.Atoi(contentLength)
			if err != nil {
				return 0, fmt.Errorf("malformed Content-Length: %s", err)
			}

			if contentLen == 0 {
				panic("chunked not implemented")
			}

			remaining := min(contentLen-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if contentLen == len(r.Body) {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("Error panic")
		}
	}

	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 8)
	readToIndex := 0
	req := &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
	for req.state != StateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != StateDone {
					return nil, fmt.Errorf("incomplete request, in state: %s, read n bytes on EOF: %d", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}
