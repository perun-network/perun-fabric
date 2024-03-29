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

package adjudicator_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"polycry.pt/poly-go/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
)

func TestValidateChannel(t *testing.T) {
	s := adjtest.NewSetup(test.Prng(t))
	require.NoError(t, adj.ValidateChannel(s.SignedChannel()))
}

func TestAdjudicator(t *testing.T) { //nolint:maintidx
	t.Run("Deposit", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(
			test.Prng(t),
			adjtest.WithChannelBalances(big.NewInt(1000), big.NewInt(1000)),
			adjtest.WithMintedTokens(big.NewInt(2000), big.NewInt(2000)),
		)

		for i := 0; i < 2; i++ {
			h, err := s.Adj.Holding(s.State.ID, s.Params.Parts[i])
			require.NoError(err)
			require.Zero(h.Sign())
		}

		th, err := s.Adj.TotalHolding(s.State.ID, s.Params.Parts)
		require.NoError(err)
		require.Zero(th.Sign())

		// Deposit twice each to test additivity.
		// As the client identification the participant address (string) is used.
		for i := 0; i < 2; i++ {
			require.NoError(s.Adj.Deposit(s.IDs[i], s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
			require.NoError(s.Adj.Deposit(s.IDs[i], s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
		}

		// Token balance for parts must be zero.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.IDs[i])
			require.Equal(big.NewInt(0), bal)
			require.NoError(err)
		}

		for i := 0; i < 2; i++ {
			h, err := s.Adj.Holding(s.State.ID, s.Params.Parts[i])
			require.NoError(err)
			doubleBal := new(big.Int).Mul(s.State.Balances[i], big.NewInt(2))
			require.Equal(doubleBal, h)
		}

		doubleTotal := new(big.Int).Mul(s.State.Total(), big.NewInt(2))
		th, err = s.Adj.TotalHolding(s.State.ID, s.Params.Parts)
		require.NoError(err)
		require.Equal(doubleTotal, th)
	})

	t.Run("Deposit-underfunded", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(
			test.Prng(t),
			adjtest.WithChannelBalances(big.NewInt(2000), big.NewInt(2000)),
			adjtest.WithMintedTokens(big.NewInt(2000), big.NewInt(1000)),
		)

		for i := 0; i < 2; i++ {
			h, err := s.Adj.Holding(s.State.ID, s.Params.Parts[i])
			require.NoError(err)
			require.Zero(h.Sign())
		}

		th, err := s.Adj.TotalHolding(s.State.ID, s.Params.Parts)
		require.NoError(err)
		require.Zero(th.Sign())

		require.NoError(s.Adj.Deposit(s.IDs[0], s.State.ID, s.Params.Parts[0], s.State.Balances[0]))
		require.Error(s.Adj.Deposit(s.IDs[1], s.State.ID, s.Params.Parts[1], s.State.Balances[1]))
	})

	t.Run("Register", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t))

		_, err := s.Adj.StateReg(s.State.ID)
		require.ErrorIs(err, adj.ErrUnknownChannel)

		sr := s.StateReg()
		ch := s.SignedChannel()
		require.NoError(s.Adj.Register(ch))

		adjsr, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr.Equal(*adjsr))
	})

	t.Run("Register-idempotence", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded, adjtest.WithVersion(2))

		sr := s.StateReg()
		ch := s.SignedChannel()
		require.NoError(s.Adj.Register(ch))
		require.NoError(s.Adj.Register(ch))

		adjsr, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr.Equal(*adjsr))
	})

	t.Run("Register-refute", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded)

		ch0 := s.SignedChannel()
		require.NoError(s.Adj.Register(ch0))

		// increment state's version
		s.State.Version = 5

		sr1 := s.StateReg()
		ch1 := s.SignedChannel()
		require.NoError(s.Adj.Register(ch1))

		adjsr1, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr1.Equal(*adjsr1))
	})

	t.Run("Register-timeout", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded)

		sr0 := s.StateReg()
		ch0 := s.SignedChannel()
		require.NoError(s.Adj.Register(ch0))

		// increment state's version
		s.State.Version = 5
		timeout := s.Ledger.Now().Add(s.Params.ChallengeDuration)
		// forward clock after dispute timeout
		s.Ledger.AdvanceNow(s.Params.ChallengeDuration + 1)

		ch1 := s.SignedChannel()
		var cterr adj.ChallengeTimeoutError
		require.ErrorAs(s.Adj.Register(ch1), &cterr)
		require.Equal(cterr.Now, s.Ledger.Now())
		require.Equal(cterr.Timeout, timeout)

		adjsr0, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr0.Equal(*adjsr0))
	})

	t.Run("Withdraw", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded)

		// Token balance for parts must be zero.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.IDs[i])
			require.Equal(big.NewInt(0), bal)
			require.NoError(err)
		}

		sr := s.StateReg()
		ch := s.SignedChannel()
		require.NoError(s.Adj.Register(ch))

		adjsr, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr.Equal(*adjsr))
		s.Ledger.AdvanceNow(s.Params.ChallengeDuration + 1)

		for i := 0; i < 2; i++ {
			req, err := adj.SignWithdrawRequest(s.Accs[i], adjsr.ID, s.IDs[i])
			require.NoError(err)
			_, err = s.Adj.Withdraw(*req)
			require.NoError(err)
		}

		// Token balance for parts must be the original value.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.IDs[i])
			require.Equal(s.State.Balances[i], bal)
			require.NoError(err)
		}
	})

	t.Run("Withdraw-invalid-sig", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded)

		// Token balance for parts must be zero.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.IDs[i])
			require.Equal(big.NewInt(0), bal)
			require.NoError(err)
		}

		sr := s.StateReg()
		ch := s.SignedChannel()
		require.NoError(s.Adj.Register(ch))

		adjsr, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.True(sr.Equal(*adjsr))
		s.Ledger.AdvanceNow(s.Params.ChallengeDuration + 1)

		// Party 0 tries to withdraw party 1's funds.
		req, err := adj.SignWithdrawRequest(s.Accs[0], adjsr.ID, s.IDs[0])
		require.NoError(err)
		req.Req.Part = s.Accs[1].Address()
		_, err = s.Adj.Withdraw(*req)
		require.Error(err)

		// Party 1 tries to withdraw party 0's funds.
		req, err = adj.SignWithdrawRequest(s.Accs[1], adjsr.ID, s.IDs[1])
		require.NoError(err)
		req.Req.Part = s.Accs[0].Address()
		_, err = s.Adj.Withdraw(*req)
		require.Error(err)

		// Check if valid withdraw still possible.
		for i := 0; i < 2; i++ {
			req, err := adj.SignWithdrawRequest(s.Accs[i], adjsr.ID, s.IDs[i])
			require.NoError(err)
			_, err = s.Adj.Withdraw(*req)
			require.NoError(err)
		}

		// Token balance for parts must be the original value.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.IDs[i])
			require.Equal(s.State.Balances[i], bal)
			require.NoError(err)
		}
	})
}
