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
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"math/big"

	"math/rand"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

const (
	challengeDuration = 10
	numParts          = 2
)

type (
	// Setup provides necessary elements for testing.
	Setup struct {
		IDs     []adj.AccountID
		Parts   []wallet.Address
		Accs    []wallet.Account
		Params  *adj.Params
		State   *adj.State
		Ledger  *TestLedger
		Adj     *adj.Adjudicator
		Timeout adj.Timestamp
	}

	// SetupOption extends the Setup constructor.
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

// NewSetup generates a new test setup.
// Each setup is created with new random accounts and an initial channel state.
func NewSetup(rng *rand.Rand, opts ...SetupOption) *Setup {
	w := wtest.NewWallet()

	var accs []wallet.Account
	// static setup options
	for _, op := range opts {
		switch staticOp := op.(type) { //nolint:gocritic
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
		ChallengeDuration: challengeDuration,
		Parts:             parts,
		Nonce:             new(big.Int).SetUint64(rng.Uint64()),
	}
	ledger := NewTestLedger()
	asset := adj.NewMemAsset()
	ids := []adj.AccountID{adj.AccountID(parts[0].String()), adj.AccountID(parts[1].String())}

	s := &Setup{
		IDs:    ids,
		Parts:  parts,
		Accs:   accs,
		Params: params,
		State: &adj.State{
			ID:       params.ID(),
			Version:  0,
			Balances: test.NewRandomBals(rng, numParts),
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

// SignedChannel returns a signed channel based on the current channel state (Params, State, Accs) of the Setup.
func (s *Setup) SignedChannel() *adj.SignedChannel {
	ch, err := adj.SignChannel(*s.Params, *s.State, s.Accs)
	if err != nil {
		panic(fmt.Sprintf("Setup: error signing channel: %v", err))
	}
	return ch.Clone()
}

// StateReg returns the StateReg according to the current state and ledger's now.
func (s *Setup) StateReg() *adj.StateReg {
	return &adj.StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone(),
	}
}

// WithFinalState allows to set the channel final flag to true.
var WithFinalState = setupModifier(func(s *Setup) {
	s.State.IsFinal = true
})

// WithVersion allows to set the state version.
func WithVersion(v uint64) SetupOption {
	return setupModifier(func(s *Setup) {
		s.State.Version = v
	})
}

// WithChannelBalances allows giving own balances instead of the default ones.
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

// WithMintedTokens mints for all participants the given amount of tokens at start.
func WithMintedTokens(fund ...*big.Int) SetupOption {
	return setupModifier(func(s *Setup) {
		if n := len(s.Params.Parts); len(fund) != n {
			panic(fmt.Sprintf(
				"Setup: error minting tokens - mismatches number of participants (%d != %d)",
				len(fund), n))
		}

		for i, id := range s.IDs {
			_ = s.Adj.Mint(id, fund[i])
		}
	})
}

// Funded performs minting and deposit for all participants.
// Therefore, the state channel is prefunded.
var Funded = setupModifier(func(s *Setup) {
	chID := s.State.ID
	for i, part := range s.Parts {
		_ = s.Adj.Mint(s.IDs[i], s.State.Balances[i])
		if err := s.Adj.Deposit(s.IDs[i], chID, part, s.State.Balances[i]); err != nil {
			panic(fmt.Sprintf("Setup: error funding participant[%d]: %v", i, err))
		}
	}
})

// WithAccounts allows setting own Accounts instead of using random ones.
func WithAccounts(accs ...wallet.Account) SetupOption {
	return withAccsOption{accs: accs}
}
