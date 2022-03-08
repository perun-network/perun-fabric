// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"encoding/json"
	"log"
	"math/big"

	"github.com/go-test/deep"
)

func BigIntWithError(b []byte, err error) (*big.Int, error) {
	if err != nil {
		return nil, err
	}

	bi := new(big.Int)
	return bi, json.Unmarshal(b, bi)
}

func RequireEqual(exp, act interface{}, msg string) {
	if diff := deep.Equal(exp, act); diff != nil {
		log.Fatalf("%s: not equal:\n%+v != %+v\nDiff:\n%v", msg, exp, act, diff)
	}
}
