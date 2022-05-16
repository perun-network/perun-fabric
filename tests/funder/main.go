package main

import (
	"context"
	"flag"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/tests"
	"log"
	pchannel "perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"polycry.pt/poly-go/test"
	"time"
)

var chainCode = flag.String("chaincode", "assetholder", "AssetHolder chaincode name")
var org = flag.Uint("org", 1, "Organization# of user to perform txs as (1 or 2)")

const testTimeout = 10 * time.Second

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	org := tests.OrgNum(*org)
	clientConn, err := tests.NewGrpcConnection(org)
	tests.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := tests.NewGateway(org, clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	ah := network.GetContract(*chainCode)
	funder := channel.NewFunder(ah)

	rng := test.Prng(test.NameStr("FabricFunder"))
	id, addr := chtest.NewRandomChannelID(rng), acc.Address()

	// Create random test parameters
	params, state := chtest.NewRandomParamsAndState(
		rng,
		chtest.WithNumAssets(1),
		chtest.WithNumParts(2),
		chtest.WithID(id),
		chtest.WithParts(addr),
	)

	log.Printf("Creating funding request ...")
	req := &pchannel.FundingReq{
		Params:    params,
		State:     state,
		Idx:       0, // TODO: Check idx
		Agreement: state.Balances,
	}

	alloc := req.State.Allocation.Sum()[0]
	log.Printf("Request created: (funding id: %v, part: %v, allocation: %v)", id, addr, alloc)

	log.Printf("Funding...")
	tests.FatalErr("funding", funder.Fund(ctx, *req))
	log.Printf("Funding successful.")
}
