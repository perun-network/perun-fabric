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
	"github.com/perun-network/perun-fabric/wallet"
	"log"
	"math/big"
	"math/rand"
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
	log.Printf("TestAdjudicatorWithSubscriptionCollaborative ...")

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	var adjs []*test.Session
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		test.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	rng := ptest.Prng(ptest.NameStr("TestAdjudicatorWithSubscriptionCollaborative"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(wallet.NewRandomAccount(rng), wallet.NewRandomAccount(rng)),
		adjtest.WithChannelBalances(big.NewInt(4000), big.NewInt(1000)))
	id := setup.State.ID

	log.Printf("Register addresses/Mint tokens...")
	for i, part := range setup.Parts {
		err := adjs[i].Binding.RegisterAddress(part)
		test.FatalErr("RegisterAddress", err)
		err = adjs[i].Binding.MintToken(part, setup.State.Balances[i])
		test.FatalErr("MintToken", err)
	}
	log.Printf("Register addresses - Successful")

	log.Printf("Depositing channel ...")
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))

		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		test.RequireEqual(bal, holding, "Holding")
		tokenBal, err := adjs[i].Binding.TokenBalance(setup.Parts[i])
		test.FatalClientErr("TokenBalance", err)
		test.RequireEqual(big.NewInt(0), tokenBal, "TokenBalance")
	}
	log.Printf("Depositing channel - Successful")
	fmt.Println("")
	subChannels := []pchannel.SignedState{} // We do not test with subchannels yet.

	// Adjudicator: Subscribe to events.
	log.Println("Subscription: Init ...")
	eventSub, err := adjs[0].Adjudicator.Subscribe(ctx, id)
	test.FatalErr("subscribe", err)
	log.Println("Subscription: Init - Successful")
	fmt.Println("")

	setup.State.IsFinal = true
	ch := setup.SignedChannel()
	req := pchannel.AdjudicatorReq{
		Params: ch.Params.CoreParams(),
		Tx: pchannel.Transaction{
			State: ch.State.CoreState(),
			Sigs:  ch.Sigs,
		},
	}

	log.Println("Register: Version 0 ...")
	// Adjudicator: Register version 0.
	{
		err := adjs[0].Adjudicator.Register(ctx, req, subChannels)
		test.FatalClientErr("register version 0", err)
	}
	log.Println("Register: Version 0 - Successful")
	fmt.Println("")

	// Adjudicator: Withdraw.
	{
		numParts := len(ch.Params.Parts)
		for _, i := range rand.Perm(numParts) {
			req.Idx = pchannel.Index(i)
			req.Acc = adjs[i].Account
			log.Printf("Withdraw: Client %d ...", req.Idx)
			err = adjs[i].Adjudicator.Withdraw(ctx, req, makeStateMapFromSignedStates(subChannels...))
			test.FatalClientErr("withdraw", err)
			tokenBal, err := adjs[i].Binding.TokenBalance(setup.Parts[i])
			test.FatalErr("TokenBalance", err)
			test.RequireEqual(setup.State.Balances[i], tokenBal, "TokenBalance")
		}
		log.Println("Withdraw - Successful")
		fmt.Println("")
	}

	// Subscription: Check concluded event.
	{
		log.Println("Subscription: Check ConcludedEvent ...")
		e, ok := eventSub.Next().(*pchannel.ConcludedEvent)
		test.RequireEqual(true, ok, "concluded")
		test.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		test.RequireEqual(e.Version() == ch.State.CoreState().Version, true, "version")
		err = e.Timeout().Wait(ctx)
		test.FatalErr("concluded: wait", err)
		log.Println("Subscription: Check ConcludedEvent - Successful")
		fmt.Println("")
	}

	// Subscription: Close.
	{
		err := eventSub.Close()
		test.FatalErr("close", err)
		err = eventSub.Err()
		test.FatalErr("err", err)
	}

	log.Printf("TestAdjudicatorWithSubscriptionCollaborative - Successful")
	fmt.Println("")
}

func withSubscriptionDispute(t *testing.T) {
	log.Printf("TestAdjudicatorWithSubscriptionDispute ...")

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
		adjtest.WithChannelBalances(big.NewInt(4000), big.NewInt(1000)))
	ch, id := setup.SignedChannel(), setup.State.ID

	log.Printf("Register addresses/Mint tokens...")
	for i, part := range setup.Parts {
		err := adjs[i].Binding.RegisterAddress(part)
		test.FatalErr("RegisterAddress", err)
		err = adjs[i].Binding.MintToken(part, setup.State.Balances[i])
		test.FatalErr("MintToken", err)
	}
	log.Printf("Register addresses - Successful")

	log.Printf("Depositing channel ...")
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		test.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, part, bal))

		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		test.RequireEqual(bal, holding, "Holding")
	}
	log.Printf("Depositing channel - Successful")
	fmt.Println("")
	subChannels := []pchannel.SignedState{} // We do not test with subchannels yet.

	// Adjudicator: Subscribe to events.
	log.Println("Subscription: Init ...")
	eventSub, err := adjs[0].Adjudicator.Subscribe(ctx, id)
	test.FatalErr("subscribe", err)
	log.Println("Subscription: Init - Successful")
	fmt.Println("")

	log.Println("Register: Version 0 ...")
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
	log.Println("Register: Version 0 - Successful")
	fmt.Println("")

	// Subscription: Check registered event version 0.
	{
		log.Println("Subscription: Check RegisteredEvent ...")
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		test.RequireEqual(true, ok, "registered")
		test.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		test.RequireEqual(e.Version() == ch.State.Version, true, "version")
		test.RequireEqual(e.State.Equal(ch.State.CoreState()) == nil, true, "equal state")
		log.Println("Subscription: Check RegisteredEvent - Successful")
		fmt.Println("")
	}

	// Adjudicator: Register version 1.
	log.Println("Register: Version 1 ...")
	setup.State.Version = 1
	setup.State.IsFinal = false
	setup.State.Balances = []pchannel.Bal{big.NewInt(3000), big.NewInt(2000)}
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
	log.Println("Register: Version 1 - Successful")
	fmt.Println("")

	// Subscription: Check registered event version 1 and wait for timeout.
	{
		log.Println("Subscription: Check RegisteredEvent ...")
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		test.RequireEqual(true, ok, "registered")
		test.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		test.RequireEqual(e.Version() == ch.State.CoreState().Version, true, "version")
		test.RequireEqual(e.State.Equal(ch.State.CoreState()) == nil, true, "equal state")
		log.Println("Subscription: Check RegisteredEvent - Successful")
		fmt.Println("")
	}

	log.Println("Check holdings ...")
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		holding, err := adjs[i].Binding.Holding(id, part)
		test.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		test.RequireEqual(bal, holding, "Holding")
	}
	log.Println("Check holdings - Successful")
	fmt.Println("")
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
		numParts := len(ch.Params.Parts)
		for _, i := range rand.Perm(numParts) {
			req.Idx = pchannel.Index(i)
			req.Acc = adjs[i].Account
			log.Printf("Withdraw: Client %d ... (takes some time because of dispute)", req.Idx)
			err = adjs[i].Adjudicator.Withdraw(ctx, req, makeStateMapFromSignedStates(subChannels...))
			test.FatalClientErr("withdraw", err)
			tokenBal, err := adjs[i].Binding.TokenBalance(setup.Parts[i])
			test.FatalErr("TokenBalance", err)
			test.RequireEqual(setup.State.Balances[i], tokenBal, "TokenBalance")
		}
		log.Println("Withdraw - Successful")
		fmt.Println("")
	}

	// Subscription: Check concluded event.
	{
		log.Println("Subscription: Check ConcludedEvent ...")
		e, ok := eventSub.Next().(*pchannel.ConcludedEvent)
		test.RequireEqual(true, ok, "concluded")
		test.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		test.RequireEqual(e.Version() == ch.State.CoreState().Version, true, "version")
		err = e.Timeout().Wait(ctx)
		test.FatalErr("concluded: wait", err)
		log.Println("Subscription: Check ConcludedEvent - Successful")
		fmt.Println("")
	}

	// Subscription: Close.
	{
		err := eventSub.Close()
		test.FatalErr("close", err)
		err = eventSub.Err()
		test.FatalErr("err", err)
	}

	log.Printf("TestAdjudicatorWithSubscriptionDispute - Successful")
	fmt.Println("")
}

func makeStateMapFromSignedStates(channels ...pchannel.SignedState) pchannel.StateMap {
	m := pchannel.MakeStateMap()
	for _, c := range channels {
		m.Add(c.State)
	}
	return m
}
