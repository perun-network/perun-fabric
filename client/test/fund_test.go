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

package client_test

import (
	"context"
	"fmt"
	"github.com/perun-network/perun-fabric/channel"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"math/big"
	"math/rand"
	pchannel "perun.network/go-perun/channel"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	simplewire "perun.network/go-perun/wire/net/simple"
	wiretest "perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
	"testing"
	"time"
)

const (
	fundTestTimeout       = 60 * time.Second
	fundChallengeDuration = 5
	fridaHolding          = 100
	fredHolding           = 100
)

func TestFundRecovery(t *testing.T) {
	rng := pkgtest.Prng(t)

	ctx, cancel := context.WithTimeout(context.Background(), fundTestTimeout)
	defer cancel()

	var (
		name  = [2]string{"Frida", "Fred"}
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
			Timeout:           10 * time.Second, // Timeout waiting for other role, not challenge duration.
			ChallengeDuration: fundChallengeDuration,
			Watcher:           watcher,
			BalanceReader:     chtest.NewBalanceReader(adjs[i].Binding, adjs[i].ClientFabricID),
		}
	}

	wiretest.SetNewRandomAccount(func(rng *rand.Rand) wire.Account { return simplewire.NewRandomAccount(rng) })
	clienttest.TestFundRecovery(
		ctx,
		t,
		clienttest.FundSetup{
			ChallengeDuration: fundChallengeDuration,
			FridaInitBal:      big.NewInt(fridaHolding),
			FredInitBal:       big.NewInt(fredHolding),
			BalanceDelta:      big.NewInt(0),
		},
		func(r *rand.Rand) ([2]clienttest.RoleSetup, pchannel.Asset) {
			return setup, channel.Asset
		},
	)
}
