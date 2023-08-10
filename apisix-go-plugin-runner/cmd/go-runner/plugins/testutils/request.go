package testutils

import (
	"context"
	"net"
	"net/http"
	"net/url"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
)

type HttpHeader struct {
	values map[string]string
}

func NewHttpHeader() *HttpHeader {
	return &HttpHeader{
		values: make(map[string]string),
	}
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key.
// The key is case insensitive
func (h *HttpHeader) Set(key string, value string) {
	h.values[key] = value
}

// Del deletes the values associated with key. The key is case insensitive
func (h *HttpHeader) Del(key string) {
	delete(h.values, key)
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case insensitive
func (h *HttpHeader) Get(key string) string {
	return h.values[key]
}

// View returns the internal structure. It is expected for read operations. Any write operation
// won't be recorded
func (h *HttpHeader) View() http.Header {
	result := http.Header(make(map[string][]string))
	for k, v := range h.values {
		result[k] = []string{v}
	}
	return result
}

type HttpRequest struct {
	ID_     uint32
	Method_ string
	Body_   []byte
	Path_   []byte
	Args_   url.Values
	Header_ pkgHTTP.Header
}

// ID returns the request id
func (r *HttpRequest) ID() uint32 {
	return r.ID_
}

// SrcIP returns the client's IP
func (r *HttpRequest) SrcIP() net.IP {
	panic("not implemented") // TODO: Implement
}

// Method returns the HTTP method (GET, POST, PUT, etc.)
func (r *HttpRequest) Method() string {
	return r.Method_
}

// Path returns the path part of the client's URI (without query string and the other parts)
// It won't be equal to the one in the Request-Line sent by the client if it has
// been rewritten by APISIX
func (r *HttpRequest) Path() []byte {
	return r.Path_
}

// SetPath is the setter for Path
func (r *HttpRequest) SetPath(path []byte) {
	r.Path_ = path
}

// Header returns the HTTP headers
func (r *HttpRequest) Header() pkgHTTP.Header {
	return r.Header_
}

// Args returns the query string
func (r *HttpRequest) Args() url.Values {
	return r.Args_
}

// Var returns the value of a Nginx variable, like `r.Var("request_time")`
//
// To fetch the value, the runner will look up the request's cache first. If not found,
// the runner will ask it from the APISIX. If the RPC call is failed, an error in
// pkg/common.ErrConnClosed type is returned.
func (r *HttpRequest) Var(name string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// Body returns HTTP request body
//
// To fetch the value, the runner will look up the request's cache first. If not found,
// the runner will ask it from the APISIX. If the RPC call is failed, an error in
// pkg/common.ErrConnClosed type is returned.
func (r *HttpRequest) Body() ([]byte, error) {
	return r.Body_, nil
}

// Context returns the request's context.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// For run plugin, the context controls cancellation.
func (r *HttpRequest) Context() context.Context {
	panic("not implemented") // TODO: Implement
}

// RespHeader returns an http.Header which allows you to add or set response headers before reaching the upstream.
// Some built-in headers would not take effect, like `connection`,`content-length`,`transfer-encoding`,`location,server`,`www-authenticate`,`content-encoding`,`content-type`,`content-location` and `content-language`
func (r *HttpRequest) RespHeader() http.Header {
	panic("not implemented") // TODO: Implement
}
