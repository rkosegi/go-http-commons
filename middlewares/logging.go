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
	"log/slog"
	"net/http"
)

var DefaultLoggingMiddleware = NewLoggingBuilder().Build()

const (
	DefHttpReqMessage  = "http request"
	DefHttpRespMessage = "http response"
)

type (
	ReqInfoExtractorFn  func(req *http.Request) (string, interface{})
	RespInfoExtractorFn func(resp InterceptedResponse) (string, interface{})
)

func HeaderReqInfoExtractor(hdr string) ReqInfoExtractorFn {
	return func(r *http.Request) (string, interface{}) {
		return hdr, r.Header.Get(hdr)
	}
}

func PathReqInfoExtractor() ReqInfoExtractorFn {
	return func(r *http.Request) (string, interface{}) {
		return "path", r.URL.Path
	}
}

func MethodReqInfoExtractor() ReqInfoExtractorFn {
	return func(r *http.Request) (string, interface{}) {
		return "method", r.Method
	}
}

func SizeReqInfoExtractor() ReqInfoExtractorFn {
	return func(r *http.Request) (string, interface{}) {
		return "size", r.ContentLength
	}
}

// DeferredReqInfoExtractor delegates extraction to provided ReqInfoExtractorFn
func DeferredReqInfoExtractor(delegate ReqInfoExtractorFn) ReqInfoExtractorFn {
	return func(r *http.Request) (string, interface{}) {
		return delegate(r)
	}
}

var DefaultReqInfoExtractors = []ReqInfoExtractorFn{
	MethodReqInfoExtractor(),
	PathReqInfoExtractor(),
	SizeReqInfoExtractor(),
}

func StatusRespInfoExtractor() RespInfoExtractorFn {
	return func(resp InterceptedResponse) (string, interface{}) {
		return "status", resp.Status()
	}
}

func SizeRespInfoExtractor() RespInfoExtractorFn {
	return func(resp InterceptedResponse) (string, interface{}) {
		return "body_size", resp.Written()
	}
}

func HeaderRespInfoExtractor(hdr string) RespInfoExtractorFn {
	return func(resp InterceptedResponse) (string, interface{}) {
		return hdr, resp.Header().Get(hdr)
	}
}

var DefaultRespInfoExtractors = []RespInfoExtractorFn{
	StatusRespInfoExtractor(),
	SizeRespInfoExtractor(),
}

// LoggingBuilder is interface to support building of logging middleware.
// slog.Logger is used as a sink.
type LoggingBuilder interface {
	// WithLogger sets slog.Logger instance to be used for logging
	WithLogger(logger *slog.Logger) LoggingBuilder

	// WithLevel sets slog.Level
	WithLevel(level slog.Level) LoggingBuilder

	// WithRequestMessage sets log message for request logging.
	// By default, it is set to DefHttpReqMessage.
	WithRequestMessage(msg string) LoggingBuilder

	// WithResponseMessage sets log message for response logging.
	// By default, it is set to DefHttpRespMessage.
	WithResponseMessage(msg string) LoggingBuilder

	// DisableResponseLog disables response logging. By default, it's enabled.
	DisableResponseLog() LoggingBuilder

	// DisableRequestLog disables request logging. By default, it's enabled.
	DisableRequestLog() LoggingBuilder

	// ClearRequestInfoExtractors clear list of ReqInfoExtractorFn.
	// By default, it holds copy of DefaultReqInfoExtractors.
	ClearRequestInfoExtractors() LoggingBuilder

	// AddRequestInfoExtractors appends provided slice of ReqInfoExtractorFn into internal slice of extractors.
	AddRequestInfoExtractors(extFns ...ReqInfoExtractorFn) LoggingBuilder

	// ClearResponseInfoExtractors clear list of RespInfoExtractorFn
	ClearResponseInfoExtractors() LoggingBuilder

	// AddResponseInfoExtractors appends provided slice of RespInfoExtractorFn into internal slice of extractors
	AddResponseInfoExtractors(extFns ...RespInfoExtractorFn) LoggingBuilder

	// Build returns MiddlewareFunc with all configured aspects.
	Build() func(http.Handler) http.Handler
}

type loggingBuilderImpl struct {
	l           *slog.Logger
	lvl         slog.Level
	reqLog      bool
	respLog     bool
	reqMsg      string
	respMsg     string
	reqInfoFns  []ReqInfoExtractorFn
	respInfoFns []RespInfoExtractorFn
}

// clone makes a copy so that modification of builder does not affect already-created MiddlewareFunc.
func (l *loggingBuilderImpl) clone() *loggingBuilderImpl {
	out := &loggingBuilderImpl{
		l:           l.l,
		lvl:         l.lvl,
		reqLog:      l.reqLog,
		respLog:     l.respLog,
		reqMsg:      l.reqMsg,
		respMsg:     l.respMsg,
		reqInfoFns:  make([]ReqInfoExtractorFn, len(l.reqInfoFns)),
		respInfoFns: make([]RespInfoExtractorFn, len(l.respInfoFns)),
	}
	copy(out.reqInfoFns, l.reqInfoFns)
	copy(out.respInfoFns, l.respInfoFns)
	return out
}

func (l *loggingBuilderImpl) ClearRequestInfoExtractors() LoggingBuilder {
	l.reqInfoFns = make([]ReqInfoExtractorFn, 0)
	return l
}

func (l *loggingBuilderImpl) AddRequestInfoExtractors(extFns ...ReqInfoExtractorFn) LoggingBuilder {
	l.reqInfoFns = append(l.reqInfoFns, extFns...)
	return l
}

func (l *loggingBuilderImpl) ClearResponseInfoExtractors() LoggingBuilder {
	l.respInfoFns = make([]RespInfoExtractorFn, 0)
	return l
}

func (l *loggingBuilderImpl) AddResponseInfoExtractors(extFns ...RespInfoExtractorFn) LoggingBuilder {
	l.respInfoFns = append(l.respInfoFns, extFns...)
	return l
}

func (l *loggingBuilderImpl) WithRequestMessage(msg string) LoggingBuilder {
	l.reqMsg = msg
	return l
}

func (l *loggingBuilderImpl) WithResponseMessage(msg string) LoggingBuilder {
	l.respMsg = msg
	return l
}

func (l *loggingBuilderImpl) DisableResponseLog() LoggingBuilder {
	l.respLog = false
	return l
}

func (l *loggingBuilderImpl) DisableRequestLog() LoggingBuilder {
	l.reqLog = false
	return l
}

func (l *loggingBuilderImpl) WithLevel(level slog.Level) LoggingBuilder {
	l.lvl = level
	return l
}

func (l *loggingBuilderImpl) WithLogger(logger *slog.Logger) LoggingBuilder {
	l.l = logger
	return l
}

func (l *loggingBuilderImpl) Build() func(http.Handler) http.Handler {
	c := l.clone()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c.reqLog {
				var args []any
				for _, fn := range c.reqInfoFns {
					k, v := fn(r)
					args = append(args, k)
					args = append(args, v)
				}
				c.l.Log(r.Context(), c.lvl, c.reqMsg, args...)
			}
			if c.respLog {
				ir := &respInterceptor{delegate: w, req: r}
				next.ServeHTTP(ir, r)
				var args []any
				for _, fn := range c.respInfoFns {
					k, v := fn(ir)
					args = append(args, k)
					args = append(args, v)
				}
				c.l.Log(r.Context(), c.lvl, c.respMsg, args...)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// NewLoggingBuilder creates LoggingBuilder with defaults
func NewLoggingBuilder() LoggingBuilder {
	return &loggingBuilderImpl{
		l:           slog.Default(),
		lvl:         slog.LevelInfo,
		respLog:     true,
		reqLog:      true,
		reqMsg:      DefHttpReqMessage,
		respMsg:     DefHttpRespMessage,
		reqInfoFns:  DefaultReqInfoExtractors,
		respInfoFns: DefaultRespInfoExtractors,
	}
}
