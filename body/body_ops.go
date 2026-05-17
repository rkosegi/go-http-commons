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

package body

import (
	"encoding/json"
	"io"
	"net/http"
)

// ConsumeAsWithDecoder consumes arbitrary request body as a given type with provided decoder function
func ConsumeAsWithDecoder[T any](req *http.Request, decFn func(io.Reader, *T) error) (*T, error) {
	var (
		res T
		err error
	)

	if err = decFn(req.Body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ConsumeAs consumes JSON request body as a given type
func ConsumeAs[T any](req *http.Request) (*T, error) {
	return ConsumeAsWithDecoder[T](req, func(r io.Reader, t *T) error {
		return json.NewDecoder(r).Decode(t)
	})
}

// PatchEntity merges existing entity with one from request body
func PatchEntity[T any, K comparable](req *http.Request, getKeyFn func(*T) K, existingSupplierFn func(K) (*T, error), mergeFn func(b1, b2 *T) (*T, error)) (*T, error) {
	fromReq, err := ConsumeAs[T](req)
	if err != nil {
		return nil, err
	}
	key := getKeyFn(fromReq)
	existing, err := existingSupplierFn(key)
	if err != nil {
		return nil, err
	}
	return mergeFn(fromReq, existing)
}
