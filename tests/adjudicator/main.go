// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
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

	clientConn, err := tests.NewGrpcConnection()
	tests.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, err := tests.NewGateway(clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	a := NewAdjudicator(network)

	rng := test.Prng(test.NameStr("FabricAdjudicator"))
	setup := adjtest.NewSetup(rng, adjtest.WithBalances(big.NewInt(4000), big.NewInt(1000)))
	ch, id := setup.SignedChannel(), setup.State.ID
	log.Printf("Depositing channel: %+v", ch)
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		tests.FatalClientErr("sending Deposit tx", a.Deposit(id, part, bal))

		holding, err := a.Holding(id, part)
		tests.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		tests.RequireEqual(bal, holding, "Holding")
	}
	log.Print("Registering state 0")
	tests.FatalClientErr("registering state 0", a.Register(ch))

	setup.State.Version = 5
	setup.State.IsFinal = true
	// transfer 1000 from participant 0 to 1
	setup.State.Balances = []channel.Bal{big.NewInt(3000), big.NewInt(2000)}
	chfinal, regfinal := setup.SignedChannel(), setup.StateReg()
	log.Print("Refuting with final state 5")
	tests.FatalClientErr("registering final state 5", a.Register(chfinal))

	log.Print("Checking final state")
	regfinal0, err := a.StateReg(id)
	tests.FatalClientErr("querying state", err)
	tests.RequireEqual(regfinal, regfinal0, "final StateReg")

	for i, part := range setup.Parts {
		withdrawn, err := a.Withdraw(id, part)
		tests.FatalClientErr("withdrawing", err)
		log.Printf("Withdrawn[%d]: %v", i, withdrawn)
		tests.RequireEqual(setup.State.Balances[i], withdrawn, "Withdraw")
	}

	log.Print("Checking that final holding is zero")
	totalfinal, err := a.TotalHolding(id, setup.Parts)
	tests.FatalClientErr("querying total holding", err)
	tests.RequireEqual(new(big.Int), totalfinal, "final zero holding")
}
