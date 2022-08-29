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
	ctest "github.com/perun-network/perun-fabric/client/test"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"

	"github.com/perun-network/perun-fabric/channel"
	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
)

const (
	happyTestTimeout  = 60 * time.Second
	happyChallengeDur = 10
	aliceHolding      = 100
	bobHolding        = 100
)

func TestHappyAliceBob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), happyTestTimeout)
	defer cancel()

	const (
		A, B = 0, 1 // Indices of Alice and Bob.
	)

	var (
		names = [2]string{"Alice", "Bob"}
		role  [2]clienttest.Executer
	)

	adjs, setup, initAssetBalance := ctest.SetupClientTest(t, names, happyChallengeDur)
	role[A] = clienttest.NewAlice(t, setup[A])
	role[B] = clienttest.NewBob(t, setup[B])

	execConfig := &clienttest.AliceBobExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			channel.Asset,
			[2]*big.Int{big.NewInt(aliceHolding), big.NewInt(bobHolding)},
			pclient.WithoutApp(),
		),
		NumPayments: [2]int{3, 2},
		TxAmounts:   [2]*big.Int{big.NewInt(10), big.NewInt(5)},
	}

	clienttest.ExecuteTwoPartyTest(ctx, t, role, execConfig)

	// Check resulting token balance.
	expectedAssetBalance := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	expectedAssetBalance[A].Sub(initAssetBalance[A], big.NewInt(20))
	expectedAssetBalance[B].Add(initAssetBalance[B], big.NewInt(20))
	for i := 0; i < len(setup); i++ {
		balance, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
		assert.NoError(t, err)
		assert.Equal(t, expectedAssetBalance[i], balance)
	}
}
