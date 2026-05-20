# httplib

`httplib` is a small Go package for building, serializing, and parsing simple HTTP/1.1-like messages.

## Installation

```bash
go get github.com/lidchen/httplib
```

## Data types

```go
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
```

## Creating values

### `NewRequest(method, path, host, body string) (*Request, error)`

Validation rules:

- `method` is required
- `path` is required and must start with `/`
- `host` is required

Behavior:

- Always initializes `Headers` with length 2
- Header 0 is always `Host`
- If `body` is non-empty, header 1 is `Content-Length`

### `NewResponse(status int, body string) (*Response, error)`

Validation rules:

- `status` must be in range `100..599`

Behavior:

- If `body` is empty, it uses `http.StatusText(status)`
- Adds `Content-Length` header based on final body

## Mutating headers

- `(*Request).WithHeader(key, value string) *Request`
- `(*Response).WithHeader(key, value string) *Response`

Both append a header and return the same pointer for chaining.

## Serialization

- `(*Request).WriteTo(io.Writer) (int64, error)`
- `(*Response).WriteTo(io.Writer) (int64, error)`
- `String()` and `MarshalText()` are implemented for both `Request` and `Response`

Current output format in `WriteTo`:

- Request first line: `METHOD PATH HTTP/1.1`
- Header lines: `Key Value` (space-separated)
- Blank line, body, trailing CRLF
- Response output includes headers + body (no HTTP status line)

## Parsing

- `ParseRequest(raw string) (Request, error)`
- `ParseResponse(raw string) (Response, error)`

Expected parse format:

- CRLF-separated lines (`\r\n`)
- Header lines must be `Key: Value`
- Request must include a `Host` header
- Request first line: `METHOD PATH HTTP/...`
- Response first line: `HTTP/... STATUS_CODE STATUS_TEXT`
- Response status text must exactly match `http.StatusText(statusCode)`

Body is parsed from lines after the first empty line and joined with `\r\n`.

## Example

```go
package main

import (
    "fmt"

    "github.com/lidchen/httplib"
)

func main() {
    req, _ := httplib.NewRequest("POST", "/api", "example.com", "hello")
    req.WithHeader("Content-Type", "text/plain")
    fmt.Println(req.String())

    raw := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nhello\r\n"
    res, _ := httplib.ParseResponse(raw)
    fmt.Println(res.StatusCode, res.Body)
}
```

## Testing

Run the package tests with:

```bash
go test ./...
```
