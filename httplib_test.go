package httplib

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewRequest_SuccessWithBody(t *testing.T) {
	req, err := NewRequest("POST", "/api", "example.com", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected method POST, got %q", req.Method)
	}
	if req.Path != "/api" {
		t.Fatalf("expected path /api, got %q", req.Path)
	}
	if req.Body != "hello" {
		t.Fatalf("expected body hello, got %q", req.Body)
	}
	if len(req.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(req.Headers))
	}
	if req.Headers[0].Key != "Host" || req.Headers[0].Value != "example.com" {
		t.Fatalf("unexpected Host header: %+v", req.Headers[0])
	}
	if req.Headers[1].Key != "Content-Length" || req.Headers[1].Value != "5" {
		t.Fatalf("unexpected Content-Length header: %+v", req.Headers[1])
	}
}

func TestNewRequest_SuccessWithoutBody(t *testing.T) {
	req, err := NewRequest("POST", "/api", "example.com", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected method POST, got %q", req.Method)
	}
	if req.Path != "/api" {
		t.Fatalf("expected path /api, got %q", req.Path)
	}
	if req.Body != "" {
		t.Fatalf("expected body empty, got %q", req.Body)
	}
	if len(req.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(req.Headers))
	}
	if req.Headers[0].Key != "Host" || req.Headers[0].Value != "example.com" {
		t.Fatalf("unexpected Host header: %+v", req.Headers[0])
	}
	if req.Headers[1].Key != "" || req.Headers[1].Value != "" {
		t.Fatalf("unexpected header: %+v", req.Headers[1])
	}
}

func TestNewResponse_SuccessWithCustomBody(t *testing.T) {
	res, err := NewResponse(201, "created")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d", res.StatusCode)
	}
	if res.Body != "created" {
		t.Fatalf("expected body created, got %q", res.Body)
	}
	if len(res.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(res.Headers))
	}
	if res.Headers[0].Key != "Content-Length" || res.Headers[0].Value != "7" {
		t.Fatalf("unexpected Content-Length header: %+v", res.Headers[0])
	}
}

func TestNewResponse_SuccessWithDefaultBody(t *testing.T) {
	res, err := NewResponse(404, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", res.StatusCode)
	}
	if res.Body != "Not Found" {
		t.Fatalf("expected default body Not Found, got %q", res.Body)
	}
	if len(res.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(res.Headers))
	}
	if res.Headers[0].Key != "Content-Length" || res.Headers[0].Value != "9" {
		t.Fatalf("unexpected Content-Length header: %+v", res.Headers[0])
	}
}

func TestNewResponse_InvalidStatus(t *testing.T) {
	invalidStatuses := []int{99, 600}

	for _, status := range invalidStatuses {
		res, err := NewResponse(status, "body")
		if err == nil {
			t.Fatalf("expected error for status %d, got nil", status)
		}
		if err.Error() != "Invalid status code" {
			t.Fatalf("unexpected error for status %d: %v", status, err)
		}
		if res != nil {
			t.Fatalf("expected nil response for status %d, got %+v", status, res)
		}
	}
}

func TestReqWithHeader(t *testing.T) {
	req, err := NewRequest("POST", "/api/v1/users", "eblog.fly.dev", `{"name": "eblog", "email": "efron.dev@gmail.com"}`)
	if err != nil {
		t.Fatalf("unexpected err for NewRequest()")
	}
	req = req.WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json").
		WithHeader("User-Agent", "httpget")
	if len(req.Headers) != 5 {
		t.Fatalf("expect 2 + 3 = 5 header with host and content-length, get %d", len(req.Headers))
	}
}

func TestResWithHeader(t *testing.T) {
	res, err := NewResponse(201, "created")
	if err != nil {
		t.Fatalf("unexpected err for NewRequest()")
	}
	res = res.WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json").
		WithHeader("User-Agent", "httpget")
	if len(res.Headers) != 4 {
		t.Fatalf("expect 1 + 3 = 4 header with body-length, get %d", len(res.Headers))
	}
}

func TestReqWriteTo(t *testing.T) {
	req, err := NewRequest("POST", "/api", "example.com", "hello")
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	req = req.WithHeader("Content-Type", "text/plain")

	var b bytes.Buffer
	n, err := req.WriteTo(&b)
	if err != nil {
		t.Fatalf("expected no error from WriteTo, got %v", err)
	}

	expected := "POST /api HTTP/1.1\r\n" +
		"Host example.com\r\n" +
		"Content-Length 5\r\n" +
		"Content-Type text/plain\r\n" +
		"\r\nhello\r\n"

	if b.String() != expected {
		t.Fatalf("unexpected request output\nexpected: %q\nactual:   %q", expected, b.String())
	}
	if n != int64(len(expected)) {
		t.Fatalf("unexpected bytes written: expected %d, got %d", len(expected), n)
	}
}

func TestResWriteTo(t *testing.T) {
	res, err := NewResponse(200, "ok")
	if err != nil {
		t.Fatalf("unexpected error creating response: %v", err)
	}
	res = res.WithHeader("Content-Type", "text/plain")

	var b bytes.Buffer
	n, err := res.WriteTo(&b)
	if err != nil {
		t.Fatalf("expected no error from WriteTo, got %v", err)
	}

	expected := "Content-Length 2\r\n" +
		"Content-Type text/plain\r\n" +
		"\r\nok\r\n"

	if b.String() != expected {
		t.Fatalf("unexpected response output\nexpected: %q\nactual:   %q", expected, b.String())
	}
	if n != int64(len(expected)) {
		t.Fatalf("unexpected bytes written: expected %d, got %d", len(expected), n)
	}
}

func TestSplitLines_EmptyString(t *testing.T) {
	got := splitLines("", "\r\n")
	if got != nil {
		t.Fatalf("expected nil for empty input, got %#v", got)
	}
}

func TestSplitLines_EmptyLine(t *testing.T) {
	got := splitLines("\r\n", "\r\n")
	if got[0] != "" {
		t.Fatalf("expected empty string for empty line, got %#v", got)
	}
}

func TestSplitLines_NoSplitter(t *testing.T) {
	got := splitLines("abc", "\r\n")
	want := []string{"abc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected output: want %#v, got %#v", want, got)
	}
}

func TestSplitLines_CRLFSeparatedLines(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("splitLines panicked: %v", r)
		}
	}()

	got := splitLines("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n", "\r\n")
	want := []string{"GET / HTTP/1.1", "Host: example.com", "", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected output: want %#v, got %#v", want, got)
	}
}

func TestParseRequest_Success(t *testing.T) {
	raw := "POST /api HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"hello\r\n"

	req, err := ParseRequest(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if req.Method != "POST" {
		t.Fatalf("expected method POST, got %q", req.Method)
	}
	if req.Path != "/api" {
		t.Fatalf("expected path /api, got %q", req.Path)
	}
	if req.Body != "hello" {
		t.Fatalf("expected body hello, got %q", req.Body)
	}

	wantHeaders := []Header{
		{Key: "Host", Value: "example.com"},
		{Key: "Content-Type", Value: "text/plain"},
	}
	if !reflect.DeepEqual(req.Headers, wantHeaders) {
		t.Fatalf("unexpected headers: want %#v, got %#v", wantHeaders, req.Headers)
	}
}

func TestParseRequest_Malformed(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "too few lines",
			raw:  "GET / HTTP/1.1\r\nHost: example.com",
		},
		{
			name: "first line missing parts",
			raw:  "GET /\r\nHost: example.com\r\n\r\n",
		},
		{
			name: "path does not start with slash",
			raw:  "GET api HTTP/1.1\r\nHost: example.com\r\n\r\n",
		},
		{
			name: "first line missing http version",
			raw:  "GET / FTP/1.0\r\nHost: example.com\r\n\r\n",
		},
		{
			name: "header missing separator",
			raw:  "GET / HTTP/1.1\r\nHost example.com\r\n\r\n",
		},
		{
			name: "missing host header",
			raw:  "GET / HTTP/1.1\r\nUser-Agent: httpget\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRequest(tt.raw)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestParseResponse_Success(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"X-Trace-Id: abc123\r\n" +
		"\r\n" +
		"hello\r\n"

	res, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("expected status code 200, got %d", res.StatusCode)
	}
	if res.Body != "hello" {
		t.Fatalf("expected body hello, got %q", res.Body)
	}

	wantHeaders := []Header{
		{Key: "Content-Type", Value: "text/plain"},
		{Key: "X-Trace-Id", Value: "abc123"},
	}
	if !reflect.DeepEqual(res.Headers, wantHeaders) {
		t.Fatalf("unexpected headers: want %#v, got %#v", wantHeaders, res.Headers)
	}
}

func TestParseResponse_Malformed(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "too few lines",
			raw:  "HTTP/1.1 200 OK\r\nContent-Type: text/plain",
		},
		{
			name: "first line missing parts",
			raw:  "HTTP/1.1 200\r\nContent-Type: text/plain\r\n\r\n",
		},
		{
			name: "missing http version",
			raw:  "FTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n",
		},
		{
			name: "status is not integer",
			raw:  "HTTP/1.1 ABC OK\r\nContent-Type: text/plain\r\n\r\n",
		},
		{
			name: "incorrect status text",
			raw:  "HTTP/1.1 200 Created\r\nContent-Type: text/plain\r\n\r\n",
		},
		{
			name: "header missing separator",
			raw:  "HTTP/1.1 200 OK\r\nContent-Type text/plain\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseResponse(tt.raw)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
