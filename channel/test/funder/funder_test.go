//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package funder_test

import (
	"context"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/test"
	"github.com/perun-network/perun-fabric/wallet"
	pchannel "perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	ptest "polycry.pt/poly-go/test"
	"testing"
	"time"
)

const (
	testTimeout = 30 * time.Second
	nrClients   = 2
)

func TestFunder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	clients := [nrClients]FunderTestClient{}
	for i := 0; i < 2; i++ {
		clients[i] = makeTestClient(uint(i+1), test.AssetholderName) // i + 1 because test org got id 1 / 2.
	}

	// Create random test parameters
	rng := ptest.Prng(ptest.NameStr("FabricFunder"))
	params, state := chtest.NewRandomParamsAndState(
		rng,
		chtest.WithNumAssets(1),
		chtest.WithParts(clients[0].acc.Address(), clients[1].acc.Address()),
	)

	// Create and send funding request.
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
	gateway, acc, _, err := test.NewGateway(org, clientConn)
	test.FatalErr("connecting to gateway", err)

	network := gateway.GetNetwork(test.ChannelName)
	funder := channel.NewFunder(network, assetholder)

	return FunderTestClient{funder: funder, acc: acc, errChan: make(chan error)}
}
