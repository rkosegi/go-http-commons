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
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/spf13/pflag"
)

var (
	ErrListenAddressMissing = errors.New("server.listen_address is required")
	ErrCorsBadMaxAge        = errors.New("invalid value of CORS max_age")
)

const (
	DefaultMetricPath = "/metrics"
)

type TLSConfig struct {
	// Path to certification bundle (cert+CAs)
	CertFile string `yaml:"cert_file"`
	// Path to private key
	KeyFile string `yaml:"key_file"`
}

// CorsConfig configures CORS
type CorsConfig struct {
	// AllowedOrigins controls value of Access-Control-Allow-Origin header.
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
	// MaxAge controls value of Access-Control-Max-Age header.
	MaxAge int `yaml:"max_age" json:"max_age"`
}

type TelemetryConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Path under which prometheus registry is exposed
	// If empty, then "/metrics" is assumed
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type ServerConfig struct {
	ListenAddress     string           `yaml:"listen_address" json:"listen_address"`
	APIPrefix         string           `yaml:"api_prefix,omitempty" json:"api_prefix,omitempty"`
	TLS               *TLSConfig       `yaml:"tls,omitempty" json:"tls,omitempty"`
	Cors              *CorsConfig      `yaml:"cors,omitempty" json:"cors,omitempty"`
	Telemetry         *TelemetryConfig `yaml:"telemetry,omitempty" json:"telemetry,omitempty"`
	ReadTimeout       *time.Duration   `yaml:"read_timeout,omitempty" json:"read_timeout,omitempty"`
	ReadHeaderTimeout *time.Duration   `yaml:"read_header_timeout,omitempty" json:"read_header_timeout,omitempty"`
	WriteTimeout      *time.Duration   `yaml:"write_timeout,omitempty" json:"write_timeout,omitempty"`
	IdleTimeout       *time.Duration   `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (t *TLSConfig) BindFlags(prefix string, pf *pflag.FlagSet) {
	pf.StringVar(&t.CertFile, prefix+"tls-cert-file", "", "TLS certificate file")
	pf.StringVar(&t.KeyFile, prefix+"tls-key-file", "", "TLS key file")
}

func (t *TelemetryConfig) BindFlags(prefix string, pf *pflag.FlagSet) {
	pf.BoolVar(&t.Enabled, prefix+"telemetry-enabled", false, "Whether to enable telemetry")
	pf.StringVar(&t.Path, prefix+"telemetry-path", "", "Telemetry path")
}

func (c *CorsConfig) BindFlags(prefix string, pf *pflag.FlagSet) {
	pf.IntVar(&c.MaxAge, prefix+"cors-max-age", c.MaxAge, "CORS MaxAge value")
	pf.StringSliceVar(&c.AllowedOrigins, prefix+"cors-allowed-origin", c.AllowedOrigins, "CORS allowed origin")
}

func (s *ServerConfig) BindFlags(prefix string, pf *pflag.FlagSet) {
	pf.StringVar(&s.ListenAddress, prefix+"listen-address", s.ListenAddress, "Address to listen on")
	pf.StringVar(&s.APIPrefix, prefix+"api-prefix", s.APIPrefix, "API prefix")
	pf.DurationVar(s.ReadTimeout, prefix+"read-timeout", 30*time.Second, "Maximum duration for reading the entire request")
	pf.DurationVar(s.ReadHeaderTimeout, prefix+"read-header-timeout", 30*time.Second, "Amount of time allowed to read request headers")
	pf.DurationVar(s.WriteTimeout, prefix+"write-timeout", 30*time.Second, "Maximum duration before timing out writes of the response")
	pf.DurationVar(s.IdleTimeout, prefix+"idle-timeout", 30*time.Second, "Maximum amount of time to wait for the next request when keep-alives are enabled")
}

// Check checks if configuration is semantically valid
func (s *ServerConfig) Check() error {
	if s.Cors != nil {
		if s.Cors.MaxAge < 0 {
			return ErrCorsBadMaxAge
		}
	}
	if s.Telemetry != nil {
		if len(s.Telemetry.Path) == 0 {
			s.Telemetry.Path = DefaultMetricPath
		}
	}
	if s.ListenAddress == "" {
		return ErrListenAddressMissing
	}
	return nil
}

func (s *ServerConfig) isTls() bool {
	if s.TLS == nil {
		return false
	}
	return len(s.TLS.CertFile) > 0 && len(s.TLS.KeyFile) > 0
}

func (s *ServerConfig) RunUntil(srv *http.Server, stopCh <-chan struct{}) error {
	var (
		err error
		l   net.Listener
	)
	if s.ReadHeaderTimeout != nil {
		srv.ReadHeaderTimeout = *s.ReadHeaderTimeout
	}
	if s.ReadTimeout != nil {
		srv.ReadTimeout = *s.ReadTimeout
	}
	if s.IdleTimeout != nil {
		srv.IdleTimeout = *s.IdleTimeout
	}
	if s.WriteTimeout != nil {
		srv.WriteTimeout = *s.WriteTimeout
	}
	if l, err = net.Listen("tcp", s.ListenAddress); err != nil {
		return err
	}
	go func() {
		if !s.isTls() {
			err = http.Serve(l, srv.Handler)
		} else {
			err = http.ServeTLS(l, srv.Handler, s.TLS.CertFile, s.TLS.KeyFile)
		}
	}()
	<-stopCh
	_ = l.Close()
	return err
}

func (s *ServerConfig) RunForever(srv *http.Server) error {
	return s.RunUntil(srv, make(chan struct{}))
}
