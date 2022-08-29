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

package test

import (
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/channel/binding"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"github.com/stretchr/testify/assert"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	simplewire "perun.network/go-perun/wire/net/simple"
	pkgtest "polycry.pt/poly-go/test"
	"testing"
	"time"
)

const clientTestTimeout = 30 * time.Second // Fixed. Not to be confused with the challenge duration.

// SetupClientTest prepares necessary objects for end-2-end testing.
// Per client a channel test session, a client role setup and the initial asset balance is returned.
func SetupClientTest(t *testing.T, name [2]string, chDuration uint64) ([]*chtest.Session, [2]clienttest.RoleSetup, [2]*big.Int) {
	t.Helper()
	rng := pkgtest.Prng(t)

	var session []*chtest.Session
	for i := uint(1); i <= 2; i++ {
		as, err := chtest.NewTestSession(chtest.OrgNum(i), chtest.AdjudicatorName)
		chtest.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		session = append(session, as)
	}

	var (
		initAssetBalance [2]*big.Int
		roleSetup        [2]clienttest.RoleSetup
	)

	bus := wire.NewLocalBus()
	for i := 0; i < len(roleSetup); i++ {
		// Build role roleSetup for test.
		watcher, _ := local.NewWatcher(session[i].Adjudicator)
		roleSetup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          simplewire.NewRandomAccount(rng),
			Bus:               bus,
			Funder:            session[i].Funder,
			Adjudicator:       session[i].Adjudicator,
			Wallet:            test.RandomWallet(),
			Timeout:           clientTestTimeout, // Timeout waiting for other role, not challenge duration.
			ChallengeDuration: chDuration,
			Watcher:           watcher,
			BalanceReader:     NewBalanceReader(session[i].Binding, session[i].ClientFabricID),
		}
		// Get current asset balances to use for checks later.
		balance, err := session[i].Binding.TokenBalance(session[i].ClientFabricID)
		assert.NoError(t, err)
		initAssetBalance[i] = balance
	}

	return session, roleSetup, initAssetBalance
}

// BalanceReader wraps the bindings TokenBalance functionality to be used in the client end-2-end tests.
type BalanceReader struct {
	binding *binding.Adjudicator
	id      adj.AccountID
}

// NewBalanceReader takes the clients binding and its fabric id to create a new BalanceReader.
func NewBalanceReader(binding *binding.Adjudicator, id adj.AccountID) *BalanceReader {
	return &BalanceReader{
		binding: binding,
		id:      id,
	}
}

// Balance returns the on-chain balance.
// We do not need a specific Asset here because we only got a single one.
func (b BalanceReader) Balance(_ pchannel.Asset) pchannel.Bal {
	balance, _ := b.binding.TokenBalance(b.id)
	return balance
}
