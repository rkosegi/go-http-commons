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

package openapi

import (
	"context"
	"log/slog"
	"net/http"
)

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type responseLoggingHttpDoer struct {
	d HttpRequestDoer
	l *slog.Logger
}

func (r *responseLoggingHttpDoer) Do(req *http.Request) (*http.Response, error) {
	resp, err := r.d.Do(req)
	if resp != nil {
		r.l.Debug("got response", "status", resp.StatusCode)
	}
	return resp, err
}

func ClientResponseLogger(d HttpRequestDoer, l *slog.Logger) HttpRequestDoer {
	return &responseLoggingHttpDoer{d: d, l: l}
}

func ClientRequestLogger(l *slog.Logger) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		l.Debug("sending request", "method", req.Method, "url", req.URL.String())
		return nil
	}
}
