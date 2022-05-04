package channel_test

import (
	"context"
	"flag"
	_ "github.com/perun-network/perun-fabric" // init backend
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/tests"
	pchannel "perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"polycry.pt/poly-go/test"
	"testing"
	"time"
)

var org = flag.Uint("org", 1, "Organization# of user to perfrom txs as (1 or 2)")

const testTimeout = 30 * time.Second

func TestFunder(t *testing.T) {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	org := tests.OrgNum(*org)
	clientConn, err := tests.NewGrpcConnection(org)
	tests.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, _, err := tests.NewGateway(org, clientConn)
	tests.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(tests.ChannelName)
	funder := channel.NewFunder(network)

	rng := test.Prng(test.NameStr("FabricFunder"))
	//id, addr := chtest.NewRandomChannelID(rng), acc.Address()

	params, state := chtest.NewRandomParamsAndState(rng, chtest.WithNumAssets(1), chtest.WithNumParts(2))
	req := newFundingRequest(params, state, 1)
	print(len(req.Agreement))
	//log.Printf("Fund: funding id: (%x, %v), amount: %v", id, addr, req.Agreement)
	tests.FatalClientErr("execute Funding tx", funder.Fund(ctx, *req))

}

// newFundingRequest returns a funding request for the specified participant and ensures that the corresponding account has sufficient funds available.
func newFundingRequest(params *pchannel.Params, state *pchannel.State, idx pchannel.Index) *pchannel.FundingReq {
	req := &pchannel.FundingReq{
		Params:    params,
		State:     state,
		Idx:       idx,
		Agreement: state.Balances,
	}

	return req
}
