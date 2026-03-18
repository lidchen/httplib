package httplib

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Header struct {
	Key, Value string
}

type Request struct {
	Method  string
	Path    string
	Headers []Header
	Body    string
}

type Response struct {
	StatusCode int
	Headers    []Header
	Body       string
}

func NewRequest(method, path, host, body string) (*Request, error) {
	if method == "" {
		return nil, errors.New("Missing requirment argument: method")
	}
	if path == "" {
		return nil, errors.New("Missing requirment argument: path")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("Path must start with '/'")
	}
	if host == "" {
		return nil, errors.New("Missing requirment argument: host")
	}
	headers := make([]Header, 2)
	headers[0] = Header{Key: "Host", Value: host}
	if body != "" {
		headers[1] = Header{Key: "Content-Length", Value: fmt.Sprintf("%d", len(body))}
	}
	return &Request{Method: method, Path: path, Headers: headers, Body: body}, nil
}

func NewResponse(status int, body string) (*Response, error) {
	if status > 599 || status < 100 {
		return nil, errors.New("Invalid status code")
	}
	if body == "" {
		body = http.StatusText(status)
	}
	header := []Header{
		{Key: "Content-Length", Value: fmt.Sprintf("%d", len(body))},
	}
	return &Response{Headers: header, StatusCode: status, Body: body}, nil
}

func (req *Request) WithHeader(key, value string) *Request {
	// TODO: Implement AsTitle(key)
	req.Headers = append(req.Headers, Header{key, value})
	return req
}

func (res *Response) WithHeader(key, value string) *Response {
	// TODO: Implement AsTitle(key)
	res.Headers = append(res.Headers, Header{key, value})
	return res
}

func (req *Request) WriteTo(w io.Writer) (n int64, err error) {
	printf := func(format string, args ...any) error {
		m, err := fmt.Fprintf(w, format, args...)
		n += int64(m)
		return err
	}
	if err := printf("%s %s HTTP/1.1\r\n", req.Method, req.Path); err != nil {
		return n, err
	}

	for _, h := range req.Headers {
		if err := printf("%s %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	err = printf("\r\n%s\r\n", req.Body)
	return n, err
}

func (res *Response) WriteTo(w io.Writer) (n int64, err error) {
	printf := func(format string, args ...any) error {
		m, err := fmt.Fprintf(w, format, args...)
		n += int64(m)
		return err
	}

	for _, h := range res.Headers {
		if err := printf("%s %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	err = printf("\r\n%s\r\n", res.Body)
	return n, err
}

// Interfaces with standard library for better experience
var _, _ fmt.Stringer = (*Request)(nil), (*Response)(nil) // compile-time check that Request and Response implement fmt.Stringer
var _, _ encoding.TextMarshaler = (*Request)(nil), (*Response)(nil)

func (req *Request) String() string  { b := new(strings.Builder); req.WriteTo(b); return b.String() }
func (res *Response) String() string { b := new(strings.Builder); res.WriteTo(b); return b.String() }
func (req *Request) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	req.WriteTo(b)
	return b.Bytes(), nil
}
func (res *Response) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	res.WriteTo(b)
	return b.Bytes(), nil
}

/* examle
GET / HTTP/1.1 \r\n
Host: localhost \r\n
User-Agent: httpget \r\n
\r\n
Body
*/

func ParseRequest(raw string) (req Request, err error) {
	errHead := "[malformed request]: "
	lines := splitLines(raw, "\r\n")
	first := strings.Fields(lines[0])
	if len(first) < 3 {
		return Request{}, errors.New(errHead + "first line should contain METHOD PATH HTTP-VERSION")
	}
	req.Method = first[0]
	req.Path = first[1]
	if !strings.HasPrefix(req.Path, "/") {
		return Request{}, errors.New(errHead + "path should begin with '/'")
	}
	if !strings.Contains(first[2], "HTTP") {
		return Request{}, errors.New(errHead + "request should contain HTTP version")
	}
	var bodyStart int
	var foundHost = false
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			// empty line, below is body part
			bodyStart = i + 1
			break
		}
		// Parsing header
		key, val, ok := strings.Cut(lines[i], ": ")
		if !ok {
			return Request{}, fmt.Errorf(errHead+"header %q should be of form 'key: value'", lines[i])
		}
		if key == "Host" {
			foundHost = true
		}
		// TODO: // key = AsTitle(key)
		req.Headers = append(req.Headers, Header{key, val})
	}
	if !foundHost {
		return Request{}, errors.New(errHead + "missing Host header")
	}
	// Parsing body
	bodyEnd := len(lines) - 1
	req.Body = strings.Join(lines[bodyStart:bodyEnd], "\r\n")
	return req, nil
}

/*
	EXAMPLE

HTTP/1.1 200 OK
Content-Length: 1300
Content-Type: text/html; charset=utf-8

body
*/
func ParseResponse(raw string) (res Response, err error) {
	errHead := "[malformed request]: "
	lines := splitLines(raw, "\r\n")
	first := strings.SplitN(lines[0], " ", 3)
	if len(first) < 3 {
		return Response{}, errors.New(errHead + "first line should contains HTTP-VERSION RESCODE RESTEXT")
	}
	if !strings.Contains(first[0], "HTTP") {
		return Response{}, errors.New(errHead + "response should contains HTTP version")
	}
	statusCode, err := strconv.Atoi(first[1])
	if err != nil {
		return Response{}, fmt.Errorf(errHead+"expect status be interger, got: %d", first[1])
	}
	// This should't happen theoretically
	if first[2] == "" || http.StatusText(statusCode) != first[2] {
		return Response{}, fmt.Errorf("missing or incorrect status text for status code %d: expected %q, but got %q", statusCode, http.StatusText(statusCode), first[2])
	}
	res.StatusCode = statusCode

	var bodyStart int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			// empty line, below is body part
			bodyStart = i + 1
			break
		}
		// Parsing header
		key, val, ok := strings.Cut(lines[i], ": ")
		if !ok {
			return Response{}, fmt.Errorf(errHead+"header %q should be of form 'key: value'", lines[i])
		}
		// TODO: // key = AsTitle(key)
		res.Headers = append(res.Headers, Header{key, val})
	}
	// Parsing body
	bodyEnd := len(lines) - 1
	res.Body = strings.Join(lines[bodyStart:bodyEnd], "\r\n")
	return res, nil
}

func splitLines(s string, spliter string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	i := 0
	for {
		j := strings.Index(s[i:], spliter)
		if j == -1 {
			lines = append(lines, s[i:])
			return lines
		}
		lines = append(lines, s[i:i+j])
		i += j + len(spliter)
	}
}
