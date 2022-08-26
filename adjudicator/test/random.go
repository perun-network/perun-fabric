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
