// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client_test //nolint:dupl

import (
	"context"
	"github.com/perun-network/perun-fabric/channel"
	ctest "github.com/perun-network/perun-fabric/client/test"
	"github.com/stretchr/testify/assert"
	"math/big"
	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	"testing"
	"time"
)

const (
	disputeTestTimeout  = 120 * time.Second
	disputeChallengeDur = 10
	malloryHolding      = 100
	carolHolding        = 100
)

func TestDisputeMalloryCarol(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), disputeTestTimeout)
	defer cancel()

	const (
		M, C = 0, 1 // Indices of Mallory and Carol.
	)

	var (
		names = [2]string{"Mallory", "Carol"}
		role  [2]clienttest.Executer
	)

	adjs, setup, initAssetBalance := ctest.SetupClientTest(t, names, disputeChallengeDur)
	role[M] = clienttest.NewMallory(t, setup[M])
	role[C] = clienttest.NewCarol(t, setup[C])

	execConfig := &clienttest.MalloryCarolExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[M].Identity.Address(), setup[C].Identity.Address()},
			channel.Asset,
			[2]*big.Int{big.NewInt(malloryHolding), big.NewInt(carolHolding)},
			pclient.WithoutApp(),
		),
		NumPayments: [2]int{3, 2},
		TxAmounts:   [2]*big.Int{big.NewInt(15), big.NewInt(10)},
	}

	clienttest.ExecuteTwoPartyTest(ctx, t, role, execConfig)

	// Check resulting token balance.
	expectedAssetBalance := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	expectedAssetBalance[M].Sub(initAssetBalance[M], big.NewInt(45)) // Only Mallory's payments expected to succeed.
	expectedAssetBalance[C].Add(initAssetBalance[C], big.NewInt(45))
	for i := 0; i < len(setup); i++ {
		balance, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
		assert.NoError(t, err)
		assert.Equal(t, expectedAssetBalance[i], balance)
	}
}
