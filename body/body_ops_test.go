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
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatchEntity(t *testing.T) {
	type Employee struct {
		Name   string
		Age    int
		Salary int
	}

	emp := Employee{Name: "Bob"}
	emp2 := Employee{Age: 25, Salary: 100}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&emp)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/employee/Bob", &buf)
	assert.NoError(t, err)

	out, err := PatchEntity[Employee](req, func(e *Employee) string {
		return e.Name
	}, func(key string) (*Employee, error) {
		return &emp2, nil
	}, func(b1, b2 *Employee) (*Employee, error) {
		return &Employee{Name: b1.Name, Age: b2.Age, Salary: b2.Salary}, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 25, out.Age)
	assert.Equal(t, 100, out.Salary)
	assert.Equal(t, "Bob", out.Name)
}
