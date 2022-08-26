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
	"github.com/perun-network/perun-fabric/channel"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"math/big"
	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	simplewire "perun.network/go-perun/wire/net/simple"
	pkgtest "polycry.pt/poly-go/test"
	"testing"
	"time"
)

const (
	disputeTestTimeout = 120 * time.Second
	malloryHolding     = 100
	carolHolding       = 100
)

func TestDisputeMalloryCarol(t *testing.T) {
	rng := pkgtest.Prng(t)

	ctx, cancel := context.WithTimeout(context.Background(), disputeTestTimeout)
	defer cancel()

	const (
		M, C = 0, 1 // Indices of Mallory and Carol
	)

	var (
		name  = [2]string{"Mallory", "Carol"}
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
}
