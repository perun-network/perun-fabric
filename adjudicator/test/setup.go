// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"math/big"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type (
	Setup struct {
		Parts   []wallet.Address
		Accs    []wallet.Account
		Params  *adj.Params
		State   *adj.State
		Ledger  *TestLedger
		Adj     *adj.Adjudicator
		Timeout adj.Timestamp
	}

	SetupOption interface {
		setupOption()
	}

	withAccsOption struct {
		accs []wallet.Account
	}

	setupModOption interface {
		SetupOption
		modify(*Setup)
	}

	setupModifier func(*Setup)
)

func (setupModifier) setupOption()      {}
func (f setupModifier) modify(s *Setup) { f(s) }

func (withAccsOption) setupOption() {}

func NewSetup(rng *rand.Rand, opts ...SetupOption) *Setup {
	w := wtest.NewWallet()

	var accs []wallet.Account
	// static setup options
	for _, op := range opts {
		switch staticOp := op.(type) {
		case withAccsOption:
			accs = staticOp.accs
		}
	}
	// generate accs if not set by option
	if accs == nil {
		accs = []wallet.Account{w.NewRandomAccount(rng), w.NewRandomAccount(rng)}
	}

	parts := []wallet.Address{accs[0].Address(), accs[1].Address()}
	params := &adj.Params{
		ChallengeDuration: 30,
		Parts:             parts,
		Nonce:             new(big.Int).SetUint64(rng.Uint64()),
	}
	ledger := NewTestLedger()
	asset := adj.NewMemAsset()
	s := &Setup{
		Parts:  parts,
		Accs:   accs,
		Params: params,
		State: &adj.State{
			ID:       params.ID(),
			Version:  0,
			Balances: test.NewRandomBals(rng, 2),
			IsFinal:  false,
		},
		Ledger:  ledger,
		Adj:     adj.NewAdjudicator(chtest.AdjudicatorName, ledger, asset),
		Timeout: ledger.Now().Add(params.ChallengeDuration),
	}

	for _, opt := range opts {
		mod, ok := opt.(setupModOption)
		if ok {
			mod.modify(s)
		}
	}

	return s
}

func (s *Setup) SignedChannel() *adj.SignedChannel {
	ch, err := adj.SignChannel(*s.Params, *s.State, s.Accs)
	if err != nil {
		panic(fmt.Sprintf("Setup: error signing channel: %v", err))
	}
	return ch.Clone()
}

// StateReg returns the StateReg according to the current state and ledger's
// now.
func (s *Setup) StateReg() *adj.StateReg {
	return &adj.StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone(),
	}
}

var WithFinalState = setupModifier(func(s *Setup) {
	s.State.IsFinal = true
})

func WithVersion(v uint64) SetupOption {
	return setupModifier(func(s *Setup) {
		s.State.Version = v
	})
}

func WithChannelBalances(bals ...channel.Bal) SetupOption {
	return setupModifier(func(s *Setup) {
		if n := len(s.Params.Parts); len(bals) != n {
			panic(fmt.Sprintf(
				"Setup: balances mismatches number of participants (%d != %d)",
				len(bals), n))
		}
		s.State.Balances = bals
	})
}

func WithMintedTokens(fund ...*big.Int) SetupOption {
	return setupModifier(func(s *Setup) {
		if n := len(s.Params.Parts); len(fund) != n {
			panic(fmt.Sprintf(
				"Setup: error minting tokens - mismatches number of participants (%d != %d)",
				len(fund), n))
		}

		for i, part := range s.Parts {
			_ = s.Adj.Mint(part.String(), fund[i])
		}
	})
}

var Funded = setupModifier(func(s *Setup) {
	id := s.State.ID
	for i, part := range s.Parts {
		_ = s.Adj.Mint(part.String(), s.State.Balances[i])
		if err := s.Adj.Deposit(part.String(), id, part, s.State.Balances[i]); err != nil {
			panic(fmt.Sprintf("Setup: error funding participant[%d]: %v", i, err))
		}
	}
})

func WithAccounts(accs ...wallet.Account) SetupOption {
	return withAccsOption{accs: accs}
}
