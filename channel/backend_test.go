// SPDX-License-Identifier: Apache-2.0

package channel_test

import (
	"testing"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
)

func TestBackend(t *testing.T) {
	rng := test.Prng(t)
	params, state := chtest.NewRandomParamsAndState(rng, chtest.WithNumLocked(int(rng.Int31n(4)+1)))
	params2, state2 := chtest.NewRandomParamsAndState(rng, chtest.WithIsFinal(!state.IsFinal), chtest.WithNumLocked(int(rng.Int31n(4)+1)))

	setup := &chtest.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wtest.NewRandomAccount(rng),
		RandomAddress: func() wallet.Address { return wtest.NewRandomAddress(rng) },
	}

	chtest.GenericBackendTest(t, setup, chtest.IgnoreApp, chtest.IgnoreAssets)
}
