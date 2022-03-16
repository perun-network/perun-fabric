// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
	"polycry.pt/poly-go/test"

	_ "perun.network/go-perun/backend/ethereum" // init ethereum backend

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
	gateway, err := newGateway(clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	a := NewAdjudicator(network)

	rng := test.Prng(test.NameStr("FabricAdjudicator"))
	setup := adjtest.NewSetup(rng)
	ch, id := setup.SignedChannel(), setup.State.ID
	log.Printf("Depositing channel: %+v", ch)
	for i, part := range setup.Params.Parts {
		bal := setup.State.Balances[i]
		tests.FatalClientErr("sending Deposit tx", a.Deposit(id, part, bal))

		holding, err := a.Holding(id, part)
		tests.FatalClientErr("querying holding", err)
		log.Printf("Querying holding: %v", holding)
		tests.RequireEqual(bal, holding, "Holding")
	}
	tests.FatalClientErr("registering state 0", a.Register(ch))

	// TODO
	// - refute with final within timeout
	// - withdraw for both users

	//withdrawn, err := ah.Withdraw(id, addr)
	//tests.FatalClientErr("withdrawing", err)
	//log.Printf("Withdrawing: %v", withdrawn)
	//tests.RequireEqual(holding, withdrawn, "Withdraw")
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
