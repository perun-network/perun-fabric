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

package assetholder_test

import (
	"github.com/perun-network/perun-fabric/channel/binding"
	"github.com/perun-network/perun-fabric/channel/test"
	"math/big"
	"testing"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	ptest "polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
)

func TestAssetholder(t *testing.T) {
	org := test.OrgNum(1)
	clientConn, err := test.NewGrpcConnection(org)
	test.FatalErr("creating client conn", err)
	defer clientConn.Close()

	// Create a Gateway connection for a specific client identity
	gateway, acc, _, err := test.NewGateway(org, clientConn)
	test.FatalErr("connecting to gateway", err)
	defer gateway.Close()

	network := gateway.GetNetwork(test.ChannelName)
	ah := binding.NewAssetHolderBinding(network, test.AssetholderName)

	rng := ptest.Prng(ptest.NameStr("FabricAssetHolder"))
	id, addr := chtest.NewRandomChannelID(rng), acc.Address()
	holding := big.NewInt(rng.Int63())
	test.FatalClientErr("sending Deposit tx", ah.Deposit(id, addr, holding))

	holding1, err := ah.Holding(id, addr)
	test.FatalClientErr("querying holding", err)
	test.RequireEqual(holding, holding1, "Holding")

	total, err := ah.TotalHolding(id, []wallet.Address{addr})
	test.FatalClientErr("querying total holding", err)
	test.RequireEqual(holding, total, "Holding")

	withdrawn, err := ah.Withdraw(id, addr)
	test.FatalClientErr("withdrawing", err)
	test.RequireEqual(holding, withdrawn, "Withdraw")

	holding2, err := ah.Holding(id, addr)
	test.FatalClientErr("querying holding", err)
	test.RequireEqual(new(big.Int), holding2, "Holding after withdrawal")
}
