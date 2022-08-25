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

package binding_test

import (
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/channel/test"
	requ "github.com/stretchr/testify/require"
	"math/big"
	"perun.network/go-perun/channel"
	ptest "polycry.pt/poly-go/test"
	"testing"

	_ "github.com/perun-network/perun-fabric" // init backend
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
)

func TestAdjudicatorBinding(t *testing.T) {
	require := requ.New(t)
	var adjs []*test.Session
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		test.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	rng := ptest.Prng(ptest.NameStr("FabricAdjudicator"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(adjs[0].Account, adjs[1].Account),
		adjtest.WithChannelBalances(big.NewInt(400), big.NewInt(100)))
	ch, id := setup.SignedChannel(), setup.State.ID

	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))
		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		require.Equal(0, bal.Cmp(holding), "Holding")
	}
	test.FatalClientErr("registering state 0 as part 0", adjs[0].Binding.Register(ch))

	setup.State.Version = 5
	setup.State.IsFinal = true
	// transfer 50 from participant 0 to 1
	setup.State.Balances = []channel.Bal{big.NewInt(350), big.NewInt(150)}
	chfinal, regfinal := setup.SignedChannel(), setup.StateReg()
	test.FatalClientErr("registering final state 5 as part 1", adjs[1].Binding.Register(chfinal))

	regfinal0, err := adjs[0].Binding.StateReg(id)
	test.FatalClientErr("querying state", err)
	require.Equal(true, regfinal.CoreState().Equal(regfinal0.CoreState()) == nil, "final StateReg")

	for i := range setup.Parts {
		req, _ := adj.SignWithdrawRequest(adjs[i].Account, setup.Params.ID(), adjs[i].ClientFabricID)
		withdrawn, err := adjs[i].Binding.Withdraw(*req)
		test.FatalClientErr("withdrawing", err)
		require.Equal(0, setup.State.Balances[i].Cmp(withdrawn), "Withdraw")
	}

	totalfinal, err := adjs[1].Binding.TotalHolding(id, setup.Parts)
	test.FatalClientErr("querying total holding", err)
	require.Equal(0, totalfinal.Cmp(new(big.Int)), "final zero holding")
}
