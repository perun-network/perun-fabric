// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log"
	"math/big"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	"polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
	"github.com/perun-network/perun-fabric/tests"
)

var chainCode = flag.String("chaincode", "assetholder", "AssetHolder chaincode name")

func main() {
	flag.Parse()

	clientConn, err := tests.NewGrpcConnection()
	tests.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := tests.NewGateway(clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	ah := NewAssetHolder(network)

	rng := test.Prng(test.NameStr("FabricAssetHolder"))
	id, addr := chtest.NewRandomChannelID(rng), acc.Address()
	holding := big.NewInt(rng.Int63())
	log.Printf("Depositing: funding id: (%x, %v), amount: %v", id, addr, holding)
	tests.FatalClientErr("sending Deposit tx", ah.Deposit(id, holding))

	holding1, err := ah.Holding(id, addr)
	tests.FatalClientErr("querying holding", err)
	log.Printf("Querying holding: %v", holding1)
	tests.RequireEqual(holding, holding1, "Holding")

	total, err := ah.TotalHolding(id, []wallet.Address{addr})
	tests.FatalClientErr("querying total holding", err)
	log.Printf("Querying total holding: %v", total)
	tests.RequireEqual(holding, total, "Holding")

	withdrawn, err := ah.Withdraw(id)
	tests.FatalClientErr("withdrawing", err)
	log.Printf("Withdrawing: %v", withdrawn)
	tests.RequireEqual(holding, withdrawn, "Withdraw")

	holding2, err := ah.Holding(id, addr)
	tests.FatalClientErr("querying holding", err)
	log.Printf("Querying holding after withdrawal: %v", holding2)
	tests.RequireEqual(new(big.Int), holding1, "Holding after withdrawal")
}
