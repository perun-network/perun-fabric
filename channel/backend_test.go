//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
