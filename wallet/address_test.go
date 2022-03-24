// SPDX-License-Identifier: Apache-2.0

package wallet_test

import (
	"crypto/elliptic"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"

	"github.com/perun-network/perun-fabric/wallet"
)

func TestAddressJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	a := wallet.NewRandomAddress(rng)
	aj, err := a.MarshalJSON()
	require.NoError(t, err)
	a0 := new(wallet.Address)
	require.NoError(t, a0.UnmarshalJSON(aj))
	require.True(t, a.Equal(a0))
}

func TestAddressBinaryMarshaling(t *testing.T) {
	rng := test.Prng(t)
	wiretest.GenericMarshalerTest(t, wallet.NewRandomAddress(rng))
}

var one = big.NewInt(1)

func incOne(x *big.Int) { x.Add(x, one) }
func decOne(x *big.Int) { x.Sub(x, one) }

func TestAddressCmp(t *testing.T) {
	rng := test.Prng(t)
	tests := []struct {
		name   string
		mod    func(*wallet.Address)
		expCmp int
	}{
		{
			name:   "=",
			mod:    func(*wallet.Address) {},
			expCmp: 0,
		},
		{
			name:   "X-1",
			mod:    func(a *wallet.Address) { decOne(a.X) },
			expCmp: -1,
		},
		{
			name:   "X+1",
			mod:    func(a *wallet.Address) { incOne(a.X) },
			expCmp: 1,
		},
		{
			name:   "Y-1",
			mod:    func(a *wallet.Address) { decOne(a.Y) },
			expCmp: -1,
		},
		{
			name:   "Y+1",
			mod:    func(a *wallet.Address) { incOne(a.Y) },
			expCmp: 1,
		},
		{
			name:   "X-1,Y-1",
			mod:    func(a *wallet.Address) { decOne(a.X); decOne(a.Y) },
			expCmp: -1,
		},
		{
			name:   "X+1,Y+1",
			mod:    func(a *wallet.Address) { incOne(a.X); incOne(a.Y) },
			expCmp: 1,
		},
		{
			name:   "X-1,Y+1",
			mod:    func(a *wallet.Address) { decOne(a.X); incOne(a.Y) },
			expCmp: -1,
		},
		{
			name:   "X+1,Y-1",
			mod:    func(a *wallet.Address) { incOne(a.X); decOne(a.Y) },
			expCmp: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := wallet.NewRandomAddress(rng)
			b := a.Clone()
			tt.mod(b)
			require.Equal(t, tt.expCmp, b.Cmp(a))
		})
	}

	t.Run("panicOnDifferentCurve", func(t *testing.T) {
		a := wallet.NewRandomAddress(rng)
		b := a.Clone()
		b.Curve = elliptic.P224() // not default P-256
		require.Panics(t, func() {
			a.Cmp(b)
		})
	})
}
