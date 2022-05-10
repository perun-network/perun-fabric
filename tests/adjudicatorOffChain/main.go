package main

import (
	"context"
	"flag"
	"fmt"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
	"github.com/perun-network/perun-fabric/tests"
	"log"
	"math/big"
	"math/rand"
	pchannel "perun.network/go-perun/channel"
	"polycry.pt/poly-go/test"
	"time"
)

var chainCode = flag.String("chaincode", "adjudicator", "AssetHolder chaincode name")

const testTimeout = 10 * time.Second

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	var adjs []*AdjudicatorSession
	for i := uint(1); i <= 2; i++ {
		as, err := NewAdjudicatorSession(tests.OrgNum(i))
		tests.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	rng := test.Prng(test.NameStr("FabricAdjudicator"))
	setup := adjtest.NewSetup(rng,
		adjtest.WithAccounts(adjs[0].Account, adjs[1].Account),
		adjtest.WithBalances(big.NewInt(4000), big.NewInt(1000)))
	ch, id := setup.SignedChannel(), setup.State.ID

	log.Printf("Depositing channel: %+v", ch)
	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		tests.FatalClientErr("sending Deposit tx", adjs[i].Binding.Deposit(id, bal))

		holding, err := adjs[i].Binding.Holding(id, part)
		tests.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		tests.RequireEqual(bal, holding, "Holding")
	}

	subChannels := []pchannel.SignedState{} // We do not test with subchannels yet.

	// Adjudicator: Subscribe to events.
	eventSub, err := adjs[0].Adjudicator.Subscribe(ctx, id)
	tests.FatalErr("subscribe", err)

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
		tests.FatalErr("register version 0", err)
	}

	// Subscription: Check registered event version 0.
	{
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		tests.RequireEqual(true, ok, "registered")
		tests.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		tests.RequireEqual(e.Version() == ch.State.Version, true, "version")
		tests.RequireEqual(e.State.Equal(ch.State.CoreState()) == nil, true, "equal state")
	}

	// Adjudicator: Register version 1.
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
		tests.FatalErr("register version 1", err)
	}

	for i, part := range setup.Parts {
		bal := setup.State.Balances[i]
		holding, err := adjs[i].Binding.Holding(id, part)
		tests.FatalClientErr("querying holding", err)
		log.Printf("Queried holding[%d]: %v", i, holding)
		tests.RequireEqual(bal, holding, "Holding")
	}

	// Subscription: Check registered event version 1 and wait for timeout.
	{
		e, ok := eventSub.Next().(*pchannel.RegisteredEvent)
		tests.RequireEqual(true, ok, "registered")
		tests.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		tests.RequireEqual(e.Version() == ch.State.CoreState().Version, true, "version")
		tests.RequireEqual(e.State.Equal(ch.State.CoreState()) == nil, true, "equal state")
		err = e.Timeout().Wait(ctx)
		tests.FatalErr("registered: wait", err)
	}

	// Adjudicator: Progress.
	// We do not test on-chain progression yet.

	// Adjudicator: Withdraw.
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
			err = adjs[i].Adjudicator.Withdraw(ctx, req, MakeStateMapFromSignedStates(subChannels...))
			tests.FatalErr("withdraw", err)
		}
	}

	// Subscription: Check concluded event.
	{
		e, ok := eventSub.Next().(*pchannel.ConcludedEvent)
		tests.RequireEqual(true, ok, "concluded")
		tests.RequireEqual(e.ID() == ch.Params.ID(), true, "equal ID")
		tests.RequireEqual(e.Version() == ch.State.CoreState().Version, true, "version")
		err = e.Timeout().Wait(ctx)
		tests.FatalErr("concluded: wait", err)
	}

	// Subscription: Close.
	{
		err := eventSub.Close()
		tests.FatalErr("close", err)
		err = eventSub.Err()
		tests.FatalErr("err", err)
	}
}

func MakeStateMapFromSignedStates(channels ...pchannel.SignedState) pchannel.StateMap {
	m := pchannel.MakeStateMap()
	for _, c := range channels {
		m.Add(c.State)
	}
	return m
}
