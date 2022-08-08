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

package client_test

import (
	"context"
	"fmt"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"github.com/stretchr/testify/assert"
	"math/big"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"testing"
	"time"

	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"

	"github.com/perun-network/perun-fabric/channel"
)

const (
	happyTestTimeout = 60 * time.Second
	aliceHolding     = 100
	bobHolding       = 100
)

func TestHappyAliceBob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), happyTestTimeout)
	defer cancel()

	const (
		A, B = 0, 1 // Indices of Alice and Bob
	)

	var (
		name  = [2]string{"Alice", "Bob"}
		role  [2]clienttest.Executer
		setup [2]clienttest.RoleSetup
	)

	var adjs []*chtest.Session
	for i := uint(1); i <= 2; i++ {
		as, err := chtest.NewTestSession(chtest.OrgNum(i), chtest.AdjudicatorName)
		chtest.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	var initAssetBalance [2]*big.Int
	bus := wire.NewLocalBus()
	for i := 0; i < len(setup); i++ {
		// Build role setup for test.
		watcher, _ := local.NewWatcher(adjs[i].Adjudicator)
		setup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          adjs[i].Account,
			Bus:               bus,
			Funder:            adjs[i].Funder,
			Adjudicator:       adjs[i].Adjudicator,
			Wallet:            test.RandomWallet(),
			Timeout:           60 * time.Second, // Timeout waiting for other role, not challenge duration.
			ChallengeDuration: 60,
			Watcher:           watcher,
		}

		// Get current asset balances to use for checks later.
		balance, err := adjs[i].Binding.TokenBalance(adjs[i].ClientID)
		assert.NoError(t, err)
		initAssetBalance[i] = balance
	}

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

	err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
	assert.NoError(t, err)

	// Check resulting token balance.
	expectedAssetBalance := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	expectedAssetBalance[0].Sub(initAssetBalance[0], big.NewInt(20))
	expectedAssetBalance[1].Add(initAssetBalance[1], big.NewInt(20))
	for i := 0; i < len(setup); i++ {
		balance, err := adjs[i].Binding.TokenBalance(adjs[i].ClientID)
		assert.NoError(t, err)
		assert.Equal(t, expectedAssetBalance[i], balance)
	}
}
