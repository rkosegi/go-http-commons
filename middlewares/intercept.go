/*
Copyright 2025 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package middlewares

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"sync"
)

// InterceptedResponse gives access to HTTP response details sent back to client
type InterceptedResponse interface {
	// Status returns response status code
	Status() int
	// Written returns number of bytes written bck to client
	Written() int
	// Header returns the header map that was sent
	Header() http.Header
	// Request returns reference to HTTP request
	Request() *http.Request
}

type respInterceptor struct {
	wroteHeader bool
	delegate    http.ResponseWriter
	written     int
	status      int
	req         *http.Request
	sendStatus  sync.Once
}

func (i *respInterceptor) Header() http.Header {
	return i.delegate.Header()
}

func (i *respInterceptor) Write(bytes []byte) (int, error) {
	i.wroteHeader = true
	size, err := i.delegate.Write(bytes)
	i.written += size
	return size, err
}

func (i *respInterceptor) WriteHeader(statusCode int) {
	if !i.wroteHeader {
		i.delegate.WriteHeader(statusCode)
		i.status = statusCode
	}
}

func (i *respInterceptor) Written() int {
	return i.written
}

func (i *respInterceptor) Status() int {
	if i.status == 0 {
		i.status = http.StatusOK
	}
	return i.status
}
func (i *respInterceptor) Request() *http.Request {
	return i.req
}

func (i *respInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := i.delegate.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

type InterceptorBuilder interface {
	// WithRequestFilter sets the custom filtering predicate.
	// If provided function return false, request is processed without interception
	WithRequestFilter(func(req *http.Request) bool) InterceptorBuilder

	WithBeforeCallback(cb func(req *http.Request)) InterceptorBuilder

	// WithCallback sets callback that will be invoked upon service request
	WithCallback(cb func(resp InterceptedResponse)) InterceptorBuilder

	// Build creates middleware function based on current builder state
	Build() func(http.Handler) http.Handler
}

type interceptorBuilderImpl struct {
	filterFn func(*http.Request) bool
	bcb      func(r *http.Request)
	cb       func(InterceptedResponse)
}

func (i *interceptorBuilderImpl) WithRequestFilter(f func(*http.Request) bool) InterceptorBuilder {
	i.filterFn = f
	return i
}

func (i *interceptorBuilderImpl) WithBeforeCallback(cb func(*http.Request)) InterceptorBuilder {
	i.bcb = cb
	return i
}

func (i *interceptorBuilderImpl) WithCallback(cb func(InterceptedResponse)) InterceptorBuilder {
	i.cb = cb
	return i
}

func (i *interceptorBuilderImpl) Build() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if i.filterFn(r) {
				i.bcb(r)
				ir := &respInterceptor{delegate: w, req: r}
				next.ServeHTTP(ir, r)
				i.cb(ir)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// NewInterceptorBuilder creates new InterceptorBuilder with the defaults.
func NewInterceptorBuilder() InterceptorBuilder {
	return &interceptorBuilderImpl{
		filterFn: func(*http.Request) bool {
			return true
		},
		bcb: func(*http.Request) {},
		cb:  func(InterceptedResponse) {},
	}
}
