// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"

	"perun.network/go-perun/channel"
	"polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
	"github.com/perun-network/perun-fabric/tests"
)

var chainCode = flag.String("chaincode", "adjudicator", "Adjudicator chaincode name")

func main() {
	flag.Parse()

	var adjs []*AdjudicatorSession
	for i := uint(1); i <= 2; i++ {
		as, err := NewAdjudicatorSession(tests.OrgNum(i))
		tests.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	rng := test.Prng(test.NameStr("FabricAdjudicator"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(adjs[0].Account, adjs[1].Account),
		adjtest.WithBalances(big.NewInt(4000), big.NewInt(1000)))
	ch, id := setup.SignedChannel(), setup.State.ID
	log.Printf("Depositing channel: %+v", ch)
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		tests.FatalClientErr("sending Deposit tx", adjs[i].Deposit(id, bal))

		holding, err := adjs[i].Holding(id, part)
		tests.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		tests.RequireEqual(bal, holding, "Holding")
	}
	log.Print("Registering state 0 as part 0")
	tests.FatalClientErr("registering state 0 as part 0", adjs[0].Register(ch))

	setup.State.Version = 5
	setup.State.IsFinal = true
	// transfer 1000 from participant 0 to 1
	setup.State.Balances = []channel.Bal{big.NewInt(3000), big.NewInt(2000)}
	chfinal, regfinal := setup.SignedChannel(), setup.StateReg()
	log.Print("Refuting with final state 5 as part 1")
	tests.FatalClientErr("registering final state 5 as part 1", adjs[1].Register(chfinal))

	log.Print("Checking final state")
	regfinal0, err := adjs[0].StateReg(id)
	tests.FatalClientErr("querying state", err)
	tests.RequireEqual(regfinal.CoreState(), regfinal0.CoreState(), "final StateReg")

	for i := range setup.Parts {
		withdrawn, err := adjs[i].Withdraw(id)
		tests.FatalClientErr("withdrawing", err)
		log.Printf("Withdrawn[%d]: %v", i, withdrawn)
		tests.RequireEqual(setup.State.Balances[i], withdrawn, "Withdraw")
	}

	log.Print("Checking that final holding is zero")
	totalfinal, err := adjs[1].TotalHolding(id, setup.Parts)
	tests.FatalClientErr("querying total holding", err)
	tests.RequireEqual(new(big.Int), totalfinal, "final zero holding")
}
