// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"math/big"
	"math/rand"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type (
	Setup struct {
		Parts  []wallet.Address
		Accs   []wallet.Account
		Params *adj.Params
		State  *adj.State
		Ledger *TestLedger
		Adj    *adj.Adjudicator
	}

	SetupOption func(*Setup)
)

func NewSetup(rng *rand.Rand, opts ...SetupOption) *Setup {
	w := wtest.NewWallet()
	accs := []wallet.Account{w.NewRandomAccount(rng), w.NewRandomAccount(rng)}
	parts := []wallet.Address{accs[0].Address(), accs[1].Address()}
	params := &adj.Params{
		ChallengeDuration: 60,
		Parts:             parts,
		Nonce:             new(big.Int).SetUint64(rng.Uint64()),
	}
	ledger := NewTestLedger()
	s := &Setup{
		Parts:  parts,
		Accs:   accs,
		Params: params,
		State: &adj.State{
			ID:       params.ID(),
			Version:  0,
			Balances: chtest.NewRandomBals(rng, 2),
			IsFinal:  false,
		},
		Ledger: ledger,
		Adj:    adj.NewAdjudicator(ledger),
	}

	for _, opt := range opts {
		opt(s)
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
		Timeout: s.Ledger.Now().Clone(),
	}
}

func Funded(s *Setup) {
	id := s.State.ID
	for i, part := range s.Parts {
		if err := s.Adj.Deposit(id, part, s.State.Balances[i]); err != nil {
			panic(fmt.Sprintf("Setup: error funding participant[%d]: %v", i, err))
		}
	}
}

func WithFinalState(s *Setup) {
	s.State.IsFinal = true
}
