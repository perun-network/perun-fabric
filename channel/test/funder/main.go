package main

import (
	"context"
	"flag"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/test"
	"github.com/perun-network/perun-fabric/wallet"
	"log"
	pchannel "perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	ptest "polycry.pt/poly-go/test"
	"time"
)

var assetholder = flag.String("assetholder", "assetholder-22618", "AssetHolder chaincode name")

//var org = flag.Uint("org", 1, "Organization# of user to perform txs as (1 or 2)")

const (
	testTimeout = 30 * time.Second
	nrClients   = 2
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	clients := [nrClients]FunderTestClient{}
	for i := 0; i < 2; i++ {
		clients[i] = makeTestClient(uint(i+1), *assetholder) // i + 1 because test org got id 1 / 2
	}

	// Create random test parameters
	rng := ptest.Prng(ptest.NameStr("FabricFunder"))
	params, state := chtest.NewRandomParamsAndState(
		rng,
		chtest.WithNumAssets(1),
		chtest.WithParts(clients[0].acc.Address(), clients[1].acc.Address()),
	)

	log.Printf("Creating and sending funding request ...")
	for i := 0; i < nrClients; i++ {
		req := &pchannel.FundingReq{
			Params:    params,
			State:     state,
			Idx:       pchannel.Index(i),
			Agreement: state.Balances,
		}
		go func(client FunderTestClient) {
			err := client.funder.Fund(ctx, *req)
			client.errChan <- err
		}(clients[i])
	}

	// Check funding
	for i := 0; i < nrClients; i++ {
		err := <-clients[i].errChan
		test.FatalErr("funding", err)
	}
	log.Printf("Funding successful.")
}

type FunderTestClient struct {
	funder  *channel.Funder
	acc     *wallet.Account
	errChan chan error
}

func makeTestClient(organization uint, assetholder string) FunderTestClient {
	org := test.OrgNum(organization)
	clientConn, err := test.NewGrpcConnection(org)
	test.FatalErr("creating client conn", err)

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := test.NewGateway(org, clientConn)
	test.FatalErr("connecting to gateway", err)

	network := gateway.GetNetwork(test.ChannelName)
	funder := channel.NewFunder(network, assetholder)

	return FunderTestClient{funder: funder, acc: acc, errChan: make(chan error)}
}
