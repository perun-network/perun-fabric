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
	"github.com/perun-network/perun-fabric/channel/test"
	"log"
	"math/big"
	"perun.network/go-perun/channel"
	ptest "polycry.pt/poly-go/test"
	"testing"

	_ "github.com/perun-network/perun-fabric" // init backend
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
)

func TestBinding(t *testing.T) {
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
		adjtest.WithChannelBalances(big.NewInt(4000), big.NewInt(1000)))
	ch, id := setup.SignedChannel(), setup.State.ID
	log.Printf("Depositing channel: %+v", ch)
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))

		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		test.RequireEqual(bal, holding, "Holding")
	}
	log.Print("Registering state 0 as part 0")
	test.FatalClientErr("registering state 0 as part 0", adjs[0].Binding.Register(ch))

	setup.State.Version = 5
	setup.State.IsFinal = true
	// transfer 1000 from participant 0 to 1
	setup.State.Balances = []channel.Bal{big.NewInt(3000), big.NewInt(2000)}
	chfinal, regfinal := setup.SignedChannel(), setup.StateReg()
	log.Print("Refuting with final state 5 as part 1")
	test.FatalClientErr("registering final state 5 as part 1", adjs[1].Binding.Register(chfinal))

	log.Print("Checking final state")
	regfinal0, err := adjs[0].Binding.StateReg(id)
	test.FatalClientErr("querying state", err)
	test.RequireEqual(regfinal.CoreState(), regfinal0.CoreState(), "final StateReg")

	for i := range setup.Parts {
		withdrawn, err := adjs[i].Binding.Withdraw(id, setup.Parts[i])
		test.FatalClientErr("withdrawing", err)
		log.Printf("Withdrawn[%d]: %v", i, withdrawn)
		test.RequireEqual(setup.State.Balances[i], withdrawn, "Withdraw")
	}

	log.Print("Checking that final holding is zero")
	totalfinal, err := adjs[1].Binding.TotalHolding(id, setup.Parts)
	test.FatalClientErr("querying total holding", err)
	test.RequireEqual(new(big.Int), totalfinal, "final zero holding")
}
