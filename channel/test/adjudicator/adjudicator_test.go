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

package adjudicator_test

import (
	"context"
	"fmt"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
	"github.com/perun-network/perun-fabric/channel/test"
	requ "github.com/stretchr/testify/require"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	ptest "polycry.pt/poly-go/test"
	"testing"
	"time"
)

const (
	testTimeout = 120 * time.Second
)

func TestAdjudicator(t *testing.T) {
	t.Run("Collaborative", withSubscriptionCollaborative)
	t.Run("Dispute", withSubscriptionDispute)
}

func withSubscriptionCollaborative(t *testing.T) {
	require := requ.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	var adjs []*test.Session
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		test.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	// Store original balances.
	var bals [2]*big.Int
	{
		for i := uint(0); i <= 1; i++ {
			bal, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
			test.FatalErr("balance", err)
			bals[i] = bal
		}
	}

	rng := ptest.Prng(ptest.NameStr("TestAdjudicatorWithSubscriptionCollaborative"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(adjs[0].Account, adjs[1].Account),
		adjtest.WithChannelBalances(big.NewInt(400), big.NewInt(100)))
	id := setup.State.ID

	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))

		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		require.Equal(0, bal.Cmp(holding), "Holding")
	}
	var subChannels []pchannel.SignedState // We do not test with subchannels yet.

	// Adjudicator: Subscribe to events.
	eventSub, err := adjs[0].Adjudicator.Subscribe(ctx, id)
	test.FatalErr("subscribe", err)

	setup.State.IsFinal = true
	ch := setup.SignedChannel()
	req := pchannel.AdjudicatorReq{
		Params: ch.Params.CoreParams(),
		Tx: pchannel.Transaction{
			State: ch.State.CoreState(),
			Sigs:  ch.Sigs,
		},
	}

	// Adjudicator: Register version 0.
	{
		err := adjs[0].Adjudicator.Register(ctx, req, subChannels)
		test.FatalClientErr("register version 0", err)
	}

	// Adjudicator: Withdraw.
	{
		for i := uint(0); i <= 1; i++ {
			req.Idx = pchannel.Index(i)
			req.Acc = adjs[i].Account
			err = adjs[i].Adjudicator.Withdraw(ctx, req, makeStateMapFromSignedStates(subChannels...))
			test.FatalClientErr("withdraw", err)
		}
	}

	// Check balances.
	{
		for i := uint(0); i <= 1; i++ {
			bal, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
			test.FatalErr("balance", err)
			require.Equal(0, bals[i].Cmp(bal), "balance not as expected")
		}
	}

	// Subscription: Check concluded event.
	{
		e, ok := eventSub.Next().(*pchannel.ConcludedEvent)
		require.Equal(true, ok, "concluded")
		require.Equal(true, e.ID() == ch.Params.ID(), "equal ID")
		require.Equal(true, e.Version() == ch.State.CoreState().Version, "version")
		err = e.Timeout().Wait(ctx)
		test.FatalErr("concluded: wait", err)
	}

	// Subscription: Close.
	{
		err := eventSub.Close()
		test.FatalErr("close", err)
		err = eventSub.Err()
		test.FatalErr("err", err)
	}
}

func withSubscriptionDispute(t *testing.T) {
	require := requ.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	var adjs []*test.Session
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		test.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	rng := ptest.Prng(ptest.NameStr("TestAdjudicatorWithSubscriptionDispute"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(adjs[0].Account, adjs[1].Account),
		adjtest.WithChannelBalances(big.NewInt(400), big.NewInt(100)))
	ch, id := setup.SignedChannel(), setup.State.ID

	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))
		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		require.Equal(0, bal.Cmp(holding), "Holding")
	}
	var subChannels []pchannel.SignedState // We do not test with subchannels yet.

	// Store balances after.
	var bals [2]*big.Int
	{
		for i := uint(0); i <= 1; i++ {
			bal, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
			test.FatalErr("balance", err)
			bals[i] = bal
		}
	}

	// Adjudicator: Subscribe to events.
	eventSub, err := adjs[0].Adjudicator.Subscribe(ctx, id)
	test.FatalErr("subscribe", err)

	// Adjudicator: Register version 0.
	{
		req := pchannel.AdjudicatorReq{
			Params: ch.Params.CoreParams(),
			Tx: pchannel.Transaction{
				State: ch.State.CoreState(),
				Sigs:  ch.Sigs,
			},
		}
		err := adjs[0].Adjudicator.Register(ctx, req, subChannels)
		test.FatalClientErr("register version 0", err)
	}

	// Subscription: Check registered event version 0.
	{
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		require.Equal(true, ok, "concluded")
		require.Equal(true, e.ID() == ch.Params.ID(), "equal ID")
		require.Equal(true, e.Version() == ch.State.CoreState().Version, "version")
		require.Equal(true, e.State.Equal(ch.State.CoreState()) == nil, "equal state")
	}

	// Adjudicator: Register version 1.
	setup.State.Version = 1
	setup.State.IsFinal = false
	setup.State.Balances = []pchannel.Bal{big.NewInt(350), big.NewInt(150)}
	ch = setup.SignedChannel()
	{
		req := pchannel.AdjudicatorReq{
			Params: ch.Params.CoreParams(),
			Tx: pchannel.Transaction{
				State: ch.State.CoreState(),
				Sigs:  ch.Sigs,
			},
		}
		err = adjs[1].Adjudicator.Register(ctx, req, subChannels)
		test.FatalClientErr("register version 1", err)
	}

	// Subscription: Check registered event version 1 and wait for timeout.
	{
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		require.Equal(true, ok, "concluded")
		require.Equal(true, e.ID() == ch.Params.ID(), "equal ID")
		require.Equal(true, e.Version() == ch.State.CoreState().Version, "version")
		require.Equal(true, e.State.Equal(ch.State.CoreState()) == nil, "equal state")
	}

	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		require.Equal(0, bal.Cmp(holding), "Holding")
	}
	// Adjudicator: Progress.
	// We do not test on-chain progression yet.

	// Adjudicator: Withdraw.
	// Withdraw can take some time (because it is waiting for conclusion)
	{
		req := pchannel.AdjudicatorReq{
			Params: ch.Params.CoreParams(),
			Tx: pchannel.Transaction{
				State: ch.State.CoreState(),
				Sigs:  ch.Sigs,
			},
		}

		for i := uint(0); i <= 1; i++ {
			req.Idx = pchannel.Index(i)
			req.Acc = adjs[i].Account
			err = adjs[i].Adjudicator.Withdraw(ctx, req, makeStateMapFromSignedStates(subChannels...))
			test.FatalClientErr("withdraw", err)
		}
	}

	// Check new balances.
	{
		for i := uint(0); i <= 1; i++ {
			bal, err := adjs[i].Binding.TokenBalance(adjs[i].ClientFabricID)
			test.FatalErr("balance", err)
			bals[i].Add(bals[i], setup.State.Balances[i])
			require.Equal(0, bal.Cmp(bals[i]), "Balance not as expected")
		}
	}

	// Subscription: Check concluded event.
	{
		e, ok := eventSub.Next().(*pchannel.ConcludedEvent)
		require.Equal(true, ok, "concluded")
		require.Equal(true, e.ID() == ch.Params.ID(), "equal ID")
		require.Equal(true, e.Version() == ch.State.CoreState().Version, "version")
		err = e.Timeout().Wait(ctx)
		test.FatalErr("concluded: wait", err)
	}

	// Subscription: Close.
	{
		err := eventSub.Close()
		test.FatalErr("close", err)
		err = eventSub.Err()
		test.FatalErr("err", err)
	}
}

func makeStateMapFromSignedStates(channels ...pchannel.SignedState) pchannel.StateMap {
	m := pchannel.MakeStateMap()
	for _, c := range channels {
		m.Add(c.State)
	}
	return m
}
