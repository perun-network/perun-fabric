// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/big"
	"math/rand"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	chtest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
)

func RandomStateReg(rng *rand.Rand, opts ...chtest.RandomOpt) *adj.StateReg {
	return &adj.StateReg{
		State:   *RandomState(rng),
		Timeout: adj.StdNow(), // random enough...
	}
}

func RandomState(rng *rand.Rand) *adj.State {
	return &adj.State{
		ID:       chtest.NewRandomChannelID(rng),
		Version:  rng.Uint64(),
		Balances: chtest.NewRandomBals(rng, 2),
		IsFinal:  rng.Int()%2 == 0,
	}
}

func RandomParams(rng *rand.Rand) *adj.Params {
	return &adj.Params{
		ChallengeDuration: rng.Uint64(),
		Parts:             RandomAddresses(rng, 2),
		Nonce:             new(big.Int).SetUint64(rng.Uint64()),
	}
}

func RandomAddresses(rng *rand.Rand, n int) []adj.Address {
	as := make([]adj.Address, 0, n)
	for i := 0; i < n; i++ {
		a := wtest.NewRandomAddress(rng).(*adj.Address)
		as = append(as, *a)
	}
	return as
}
