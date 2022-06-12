package main

import (
	"context"
	"flag"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/test"
	"log"
	pchannel "perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	ptest "polycry.pt/poly-go/test"
	"time"
)

var assetholder = flag.String("assetholder", "assetholder-8195", "AssetHolder chaincode name")
var org = flag.Uint("org", 1, "Organization# of user to perform txs as (1 or 2)")

const testTimeout = 10 * time.Second

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	org := test.OrgNum(*org)
	clientConn, err := test.NewGrpcConnection(org)
	test.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := test.NewGateway(org, clientConn)
	test.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(test.ChannelName)
	funder := channel.NewFunder(network, *assetholder)

	rng := ptest.Prng(ptest.NameStr("FabricFunder"))
	addr := acc.Address()

	// Create random test parameters
	params, state := chtest.NewRandomParamsAndState(
		rng,
		chtest.WithNumAssets(1),
		chtest.WithNumParts(2),
		chtest.WithParts(addr),
	)

	log.Printf("Creating funding request ...")
	req := &pchannel.FundingReq{
		Params:    params,
		State:     state,
		Idx:       0,
		Agreement: state.Balances,
	}

	alloc := req.State.Allocation.Sum()[0]
	log.Printf("Request created: (part: %v, allocation: %v)", addr, alloc)

	log.Printf("Funding...")
	test.FatalErr("funding", funder.Fund(ctx, *req))
	log.Printf("Funding successful.")
}
