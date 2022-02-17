// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/rand"

	"github.com/perun-network/perun-fabric/adjudicator"
	chtest "perun.network/go-perun/channel/test"
)

func RandomStateReg(rng *rand.Rand, opts ...chtest.RandomOpt) *adjudicator.StateReg {
	return &adjudicator.StateReg{
		State:   chtest.NewRandomState(rng, opts...),
		Timeout: adjudicator.StdNow(), // random enough...
	}
}
