// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/perun-network/perun-fabric/pkg/json"
	"github.com/stretchr/testify/require"
)

type errMarshaler struct{}

func (errMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("doh")
}

func TestMultiMarshal(t *testing.T) {
	tests := []struct {
		name   string
		xs     []interface{}
		jsons  []string
		hasErr bool
	}{
		{
			name:  "empty",
			xs:    nil,
			jsons: nil,
		},
		{
			name:  "one",
			xs:    []interface{}{8},
			jsons: []string{"8"},
		},
		{
			name: "multi",
			xs: []interface{}{8, big.NewInt(42), "perun", struct {
				X int
				Y string
			}{X: 23, Y: "poly"}},
			jsons: []string{"8", "42", `"perun"`, `{"X":23,"Y":"poly"}`},
		},
		{
			name:   "multi-err",
			xs:     []interface{}{8, big.NewInt(42), "perun", errMarshaler{}},
			jsons:  []string{"8", "42", `"perun"`},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			js, err := json.MultiMarshal(test.xs...)
			if test.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.jsons, js)
		})
	}
}
