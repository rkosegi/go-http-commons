/*
Copyright 2026 Richard Kosegi

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

package servertypes

import (
	"context"
	"io"
)

// RunCloser is implemented by servers that could be closed
type RunCloser interface {
	io.Closer
	// Run runs a server and blocks caller until given context is done.
	Run(ctx context.Context) error
}

// SystemVersionInfo encapsulates common version information
type SystemVersionInfo struct {
	// BuildTime Time when application was built
	BuildTime *string `json:"build-time,omitempty" yaml:"build-time,omitempty"`

	// BuildUser User who built application
	BuildUser *string `json:"build-user,omitempty" yaml:"build-user,omitempty"`

	// Revision VCS revision
	Revision *string `json:"revision,omitempty" yaml:"revision,omitempty"`

	// Version Application version
	Version *string `json:"version,omitempty" yaml:"version,omitempty"`
}
