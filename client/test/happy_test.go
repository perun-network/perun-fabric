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
	"math/big"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/watcher/local"
	ptest "polycry.pt/poly-go/test"
	"testing"
	"time"

	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"

	"github.com/perun-network/perun-fabric/channel"
	"github.com/stretchr/testify/assert"
)

const (
	happyTestTimeout = 60 * time.Second
	aliceHolding     = 1000
	bobHolding       = 1000
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

	bus := wire.NewLocalBus()
	var addrs [2]wallet.Address
	for i := 0; i < len(setup); i++ {
		watcher, _ := local.NewWatcher(adjs[i].Adjudicator)

		// We need the address for minting the tokens to be the same as the participant in the channel.
		// ConstWallet implements a special behavior that allows us to ensure this.
		rng := ptest.Prng(ptest.NameStr(fmt.Sprintf("HappyTestClient%d", i)))
		constWallet := chtest.NewConstTestWallet(rng, 1)

		// Store address for balance check later.
		addr := constWallet.GetConstAccount().Address()
		addrs[i] = addr

		// Mint tokens.
		err := adjs[i].Binding.RegisterAddress(addr)
		assert.NoError(t, err)
		err = adjs[i].Binding.MintToken(addr, big.NewInt(1000))
		assert.NoError(t, err)

		setup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          adjs[i].Account,
			Bus:               bus,
			Funder:            adjs[i].Funder,
			Adjudicator:       adjs[i].Adjudicator,
			Wallet:            constWallet,
			Timeout:           60 * time.Second, // Timeout waiting for other role, not challenge duration.
			ChallengeDuration: 60,
			Watcher:           watcher,
		}
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
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(50), big.NewInt(0)},
	}

	err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
	assert.NoError(t, err)

	// Check resulting token balance.
	expected := [2]*big.Int{big.NewInt(750), big.NewInt(1250)}
	for i := 0; i < len(setup); i++ {
		balance, err := adjs[i].Binding.TokenBalance(addrs[i])
		assert.NoError(t, err)
		assert.Equal(t, expected[i], balance)
	}
}
