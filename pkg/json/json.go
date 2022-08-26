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

package json

import (
	"encoding/json"
	"fmt"
)

// MultiMarshal calls json.Marshal on all xs and returns a slice of all
// resulting strings. If any marshaling operation fails, it returns that error
// together with the marshalings generated up to that point.
func MultiMarshal(xs ...interface{}) ([]string, error) {
	var js []string //nolint:prealloc
	for i, x := range xs {
		j, err := json.Marshal(x)
		if err != nil {
			return js, fmt.Errorf("JSON-marshaling element %d of type %T: %v", i, x, err)
		}
		js = append(js, string(j))
	}
	return js, nil
}
