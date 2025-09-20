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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigCheck(t *testing.T) {
	var c *ServerConfig
	t.Run("CORS max age invalid", func(t *testing.T) {
		c = &ServerConfig{Cors: &CorsConfig{MaxAge: -1}}
		assert.Error(t, c.Check())
	})
	t.Run("is TLS enabled", func(t *testing.T) {
		c = &ServerConfig{TLS: &TLSConfig{}, ListenAddress: ":8080"}
		assert.False(t, c.isTls())
		c.TLS.CertFile = "cert.pem"
		assert.False(t, c.isTls())
		c.TLS.KeyFile = "key.pem"
		assert.True(t, c.isTls())
	})
	t.Run("Default telemetry path", func(t *testing.T) {
		c = &ServerConfig{Telemetry: &TelemetryConfig{Enabled: true}, ListenAddress: ":8080"}
		assert.NoError(t, c.Check())
		assert.Equal(t, DefaultMetricPath, c.Telemetry.Path)
	})
}
