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
	"fmt"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"math/big"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"testing"
	"time"

	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	simplewire "perun.network/go-perun/wire/net/simple"

	"github.com/perun-network/perun-fabric/channel"
	pkgtest "polycry.pt/poly-go/test"
)

const (
	happyTestTimeout = 60 * time.Second
	aliceHolding     = 100
	bobHolding       = 100
)

func TestHappyAliceBob(t *testing.T) {
	rng := pkgtest.Prng(t)

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
	for i := 0; i < len(setup); i++ {
		// Build role setup for test.
		watcher, _ := local.NewWatcher(adjs[i].Adjudicator)
		setup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          simplewire.NewRandomAccount(rng),
			Bus:               bus,
			Funder:            adjs[i].Funder,
			Adjudicator:       adjs[i].Adjudicator,
			Wallet:            test.RandomWallet(),
			Timeout:           30 * time.Second, // Timeout waiting for other role, not challenge duration.
			ChallengeDuration: 10,
			Watcher:           watcher,
			BalanceReader:     chtest.NewBalanceReader(adjs[i].Binding, adjs[i].ClientFabricID),
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
		NumPayments: [2]int{3, 2},
		TxAmounts:   [2]*big.Int{big.NewInt(10), big.NewInt(5)},
	}

	clienttest.ExecuteTwoPartyTest(ctx, t, role, execConfig)
}
