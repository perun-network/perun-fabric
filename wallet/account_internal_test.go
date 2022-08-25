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

package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"polycry.pt/poly-go/test"
)

var curves = []elliptic.Curve{
	elliptic.P224(),
	elliptic.P256(),
	elliptic.P384(),
	elliptic.P521(),
}

func TestSigMarshalingDecoding(t *testing.T) {
	rng := test.Prng(t)
	for i := 0; i < 0x20; i++ {
		curve := curves[rng.Int()%len(curves)]
		sk, err := ecdsa.GenerateKey(curve, rng)
		require.NoError(t, err)

		// any hash will do...
		digest := Hash(sk.X.Bytes())
		r, s, err := ecdsa.Sign(rng, sk, digest)
		require.NoError(t, err)
		sig := marshalSig(curve, r, s)

		t.Run(fmt.Sprintf("unmarshaling-%d", i), func(t *testing.T) {
			r0, s0, err := unmarshalSig(sig)
			require.NoError(t, err)
			require.Equal(t, r, r0)
			require.Equal(t, s, s0)
		})

		t.Run(fmt.Sprintf("decoding-%d", i), func(t *testing.T) {
			reader := bytes.NewBuffer(sig)
			sig0, err := DecodeSig(reader)
			require.NoError(t, err)
			require.Equal(t, sig, sig0)
		})

		t.Run(fmt.Sprintf("decoding-short-err-%d", i), func(t *testing.T) {
			reader := bytes.NewBuffer(sig[:len(sig)-1])
			_, err := DecodeSig(reader)
			require.Error(t, err)
		})
	}
}
