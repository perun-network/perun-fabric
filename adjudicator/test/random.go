// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/big"
	"math/rand"

	chtest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

// RandomStateReg returns a random state registration for testing.
func RandomStateReg(rng *rand.Rand, opts ...chtest.RandomOpt) *adj.StateReg {
	return &adj.StateReg{
		State:   *RandomState(rng),
		Timeout: adj.StdNow(), // random enough...
	}
}

// RandomState returns a random channel state for testing.
func RandomState(rng *rand.Rand) *adj.State {
	return &adj.State{
		ID:       chtest.NewRandomChannelID(rng),
		Version:  rng.Uint64(),
		Balances: chtest.NewRandomBals(rng, numParts),
		IsFinal:  rng.Int()%2 == 0,
	}
}

// RandomParams returns random channel parameters for testing.
func RandomParams(rng *rand.Rand) *adj.Params {
	return &adj.Params{
		ChallengeDuration: rng.Uint64(),
		Parts:             wtest.NewRandomAddresses(rng, numParts),
		Nonce:             new(big.Int).SetUint64(rng.Uint64()),
	}
}
