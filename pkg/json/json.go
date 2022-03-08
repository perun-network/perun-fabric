package json

import (
	"encoding/json"
	"fmt"
)

// MultiMarshal calls json.Marshal on all xs and returns a slice of all
// resulting strings. If any marshaling operation fails, it returns that error
// together with the marshalings generated up to that point.
func MultiMarshal(xs ...interface{}) (js []string, _ error) {
	for i, x := range xs {
		j, err := json.Marshal(x)
		if err != nil {
			return js, fmt.Errorf("JSON-marshaling element %d of type %T: %v", i, x, err)
		}
		js = append(js, string(j))
	}
	return js, nil
}
