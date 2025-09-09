package request

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"httpfromtcp/internal/headers"
)

type RequestState int

const (
	Initialized RequestState = iota
	ParsingHeaders
	ParsingBody
	Done
)

var stateName = map[RequestState]string{
	Initialized:    "initialized",
	ParsingHeaders: "parsingHeaders",
	ParsingBody:    "parsingBody",
	Done:           "done",
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const SEPERATOR = "\r\n"
const BUFFER_SIZE = 1024

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, BUFFER_SIZE)
	readToIndex := 0
	r := &Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
	}

	for r.state != Done {
		if len(buff) == readToIndex {
			newBuff := make([]byte, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}

		n, err := reader.Read(buff[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.state = Done
				break
			}
			return nil, err
		}
		readToIndex += n

		n = math.MaxInt
		for n != 0 {
			n, err = r.parse(buff[:readToIndex])
			if err != nil {
				return nil, err
			}

			if r.state == Done {
				break
			}

			copy(buff, buff[n:readToIndex])
			readToIndex -= n
		}
	}

	cl, err := r.Headers.GetContentLength()
	if err != nil {
		return nil, err
	}
	if len(r.Body) < cl {
		return r, fmt.Errorf("content smaller than content-length")
	}

	return r, nil
}

func parseRequestLine(s string) (*RequestLine, int, error) {
	i := strings.Index(s, SEPERATOR)
	if i == -1 {
		return nil, 0, nil
	}

	s = s[:i]
	consumed := len(s) + len(SEPERATOR)

	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, consumed, fmt.Errorf("malformed request line")
	}

	method := parts[0]
	for _, c := range method {
		if !unicode.IsUpper(c) {
			return nil, consumed, fmt.Errorf("method is not all Upper. Problamatic rune: %c", c)
		}
		if !unicode.IsLetter(c) {
			return nil, consumed, fmt.Errorf("method is not all Alphanumeric. Problamatic rune: %c", c)
		}
	}

	httpVersion := parts[2]
	if !strings.Contains(httpVersion, "1.1") {
		return nil, consumed, fmt.Errorf("wrong http version. only 1.1 is supported")
	}
	httpVersion = strings.Split(httpVersion, "/")[1]
	requestLine := RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: parts[1],
		Method:        method,
	}
	return &requestLine, consumed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case Done:
		return 0, fmt.Errorf("trying to read data in a done state")
	case Initialized:
		return r.parseRequestLine(data)
	case ParsingHeaders:
		return r.parseHeaders(data)
	case ParsingBody:
		return r.parseBody(data)
	default:
		return 0, fmt.Errorf("unreachable state: %v", r.state)
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	rl, n, err := parseRequestLine(string(data))
	if n == 0 && err == nil {
		return 0, nil
	}
	if err != nil {
		return n, err
	}

	r.RequestLine = *rl
	r.state = ParsingHeaders
	return n, nil
}

func (r *Request) parseHeaders(data []byte) (int, error) {
	n, done, err := r.Headers.Parse(data)
	if err != nil {
		return n, err
	}
	if done {
		r.state = ParsingBody
	}
	return n, nil
}

func (r *Request) parseBody(data []byte) (int, error) {
	contentLength := r.Headers.Get("content-length")
	if contentLength == "" {
		r.state = Done
		return 0, nil
	}

	cl, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}

	remaining := cl - len(r.Body)
	if remaining <= 0 {
		r.state = Done
		return 0, nil
	}

	toCopy := min(len(data), remaining)
	r.Body = append(r.Body, data[:toCopy]...)

	if len(r.Body) == cl {
		r.state = Done
	}

	return toCopy, nil
}
