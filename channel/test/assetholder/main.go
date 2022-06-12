// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/perun-network/perun-fabric/channel/binding"
	"github.com/perun-network/perun-fabric/channel/test"
	"log"
	"math/big"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	ptest "polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
)

var assetholder = flag.String("assetholder", "assetholder-8195", "AssetHolder chaincode name")
var org = flag.Uint("org", 1, "Organization# of user to perform txs as (1 or 2)")

func main() {
	flag.Parse()

	org := test.OrgNum(*org)
	clientConn, err := test.NewGrpcConnection(org)
	test.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := test.NewGateway(org, clientConn)
	test.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(test.ChannelName)
	ah := binding.NewAssetHolderBinding(network, *assetholder)

	rng := ptest.Prng(ptest.NameStr("FabricAssetHolder"))
	id, addr := chtest.NewRandomChannelID(rng), acc.Address()
	holding := big.NewInt(rng.Int63())
	log.Printf("Depositing: funding id: (%x, %v), amount: %v", id, addr, holding)
	test.FatalClientErr("sending Deposit tx", ah.Deposit(id, holding))

	holding1, err := ah.Holding(id, addr)
	test.FatalClientErr("querying holding", err)
	log.Printf("Querying holding: %v", holding1)
	test.RequireEqual(holding, holding1, "Holding")

	total, err := ah.TotalHolding(id, []wallet.Address{addr})
	test.FatalClientErr("querying total holding", err)
	log.Printf("Querying total holding: %v", total)
	test.RequireEqual(holding, total, "Holding")

	withdrawn, err := ah.Withdraw(id)
	test.FatalClientErr("withdrawing", err)
	log.Printf("Withdrawing: %v", withdrawn)
	test.RequireEqual(holding, withdrawn, "Withdraw")

	holding2, err := ah.Holding(id, addr)
	test.FatalClientErr("querying holding", err)
	log.Printf("Querying holding after withdrawal: %v", holding2)
	test.RequireEqual(new(big.Int), holding1, "Holding after withdrawal")
}
