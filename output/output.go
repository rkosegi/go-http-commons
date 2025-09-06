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

package output

import (
	"encoding/json"
	"io"
	"net/http"
)

var defaultOutput = NewBuilder().Build()

func DefaultOutput() Interface {
	return defaultOutput
}

// ErrorMapper is function that could customize serialization of error instance.
// All registered mappers all enumerated in order they were registered when instance of error is to be sent.
// First mapper that returns true will terminate enumeration and process will end.
type ErrorMapper func(w http.ResponseWriter, err error) bool

// PayloadEncoder is responsible to serialize object to io.Writer
type PayloadEncoder func(w io.Writer, objToWrite interface{}) error

type Interface interface {
	// SendWithStatus sends object back to client with given status code.
	// If object to be send is error, then registered ErrorMappers can influence output instead.
	SendWithStatus(w http.ResponseWriter, v interface{}, status int)
	// SendBytes sends raw bytes to output, assuming content-type of current encoder
	SendBytes(w http.ResponseWriter, data []byte)
}

type Builder interface {
	// WithEncoder sets default content-type and PayloadEncoder that serializes
	// objects into io.Writer in that content-type.
	WithEncoder(ct string, encFunc PayloadEncoder) Builder
	// WithErrorMapper registers mapping function that can customize error response
	WithErrorMapper(mapper ErrorMapper) Builder
	// Build creates new Interface using current state of this Builder.
	Build() Interface
}

func defaultJsonEncoder() PayloadEncoder {
	return func(w io.Writer, v interface{}) error {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		return e.Encode(v)
	}
}

type builder struct {
	encFn PayloadEncoder
	ct    string
	ems   []ErrorMapper
}

func (b *builder) WithErrorMapper(mapper ErrorMapper) Builder {
	b.ems = append(b.ems, mapper)
	return b
}

type impl struct {
	*builder
}

func (i *impl) SendBytes(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", i.ct)
	_, _ = w.Write(data)
}

func (i *impl) SendWithStatus(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", i.ct)
	w.WriteHeader(status)
	// check if value to send is error
	if err, ok := v.(error); ok {
		// if so, consult the list of ErrorMappers, if one of them can handle that
		for _, em := range i.ems {
			if em(w, err) {
				return
			}
		}
		// otherwise use http.Error
		http.Error(w, err.Error(), status)
		return
	}
	// encode value using configured PayloadEncoder
	if err := i.encFn(w, v); err != nil {
		// last resort to send back something
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (b *builder) Build() Interface {
	return &impl{
		builder: &builder{
			encFn: b.encFn,
			ct:    b.ct,
			ems:   b.ems,
		},
	}
}

func (b *builder) WithEncoder(ct string, encFunc PayloadEncoder) Builder {
	b.ct = ct
	b.encFn = encFunc
	return b
}

func NewBuilder() Builder {
	return &builder{
		encFn: defaultJsonEncoder(),
		ct:    "application/json",
	}
}
