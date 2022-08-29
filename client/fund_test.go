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
	"github.com/perun-network/perun-fabric/channel"
	ctest "github.com/perun-network/perun-fabric/client/test"
	"math/big"
	"math/rand"
	pchannel "perun.network/go-perun/channel"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	simplewire "perun.network/go-perun/wire/net/simple"
	wiretest "perun.network/go-perun/wire/test"
	"testing"
	"time"
)

const (
	fundTestTimeout       = 120 * time.Second
	fundChallengeDuration = 5
	fridaHolding          = 100
	fredHolding           = 100
)

func TestFundRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), fundTestTimeout)
	defer cancel()

	var (
		names = [2]string{"Frida", "Fred"}
		setup [2]clienttest.RoleSetup
	)

	_, setup, _ = ctest.SetupClientTest(t, names, disputeChallengeDur) // Adjudicator and initial bals not needed.
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
