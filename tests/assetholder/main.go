// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
	chtest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"

	_ "perun.network/go-perun/backend/ethereum" // init sim backend

	"github.com/perun-network/perun-fabric/tests"
)

var chainCode = flag.String("chaincode", "assetholder", "AssetHolder chaincode name")

func main() {
	flag.Parse()

	clientConn, err := tests.NewGrpcConnection()
	tests.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, err := newGateway(clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	ah := NewAssetHolder(network)

	rng := test.Prng(test.NameStr("FabricAssetHolder"))
	id, addr := chtest.NewRandomChannelID(rng), wtest.NewRandomAddress(rng)
	holding := big.NewInt(rng.Int63())
	log.Printf("Depositing: funding id: (%x, %v), amount: %v", id, addr, holding)
	tests.FatalClientErr("sending Deposit tx", ah.Deposit(id, addr, holding))

	holding1, err := ah.Holding(id, addr)
	tests.FatalClientErr("querying holding", err)
	log.Printf("Querying holding: %v", holding1)
	tests.RequireEqual(holding, holding1, "Holding")

	withdrawn, err := ah.Withdraw(id, addr)
	tests.FatalClientErr("withdrawing", err)
	log.Printf("Withdrawing: %v", withdrawn)
	tests.RequireEqual(holding, withdrawn, "Withdraw")

	holding2, err := ah.Holding(id, addr)
	tests.FatalClientErr("querying holding", err)
	log.Printf("Querying holding after withdrawal: %v", holding2)
	tests.RequireEqual(new(big.Int), holding1, "Holding after withdrawal")
}

func newGateway(clientConn *grpc.ClientConn) (*client.Gateway, error) {
	id, err := tests.NewIdentity()
	if err != nil {
		return nil, err
	}
	sign, err := tests.NewSign()
	if err != nil {
		return nil, err
	}
	tests.FatalErr("creating signer", err)

	// Create a Gateway connection for a specific client identity
	return client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConn),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
}
