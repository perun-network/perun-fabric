// SPDX-License-Identifier: Apache-2.0

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

func TestAdjudicator(t *testing.T) {
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
		// partID in the local case not necessary.
		for i := 0; i < 2; i++ {
			require.NoError(s.Adj.Deposit(s.Params.Parts[i].String(), s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
			require.NoError(s.Adj.Deposit(s.Params.Parts[i].String(), s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
		}

		// Token balance for parts must be zero.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.Params.Parts[i].String())
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

		require.NoError(s.Adj.Deposit(s.Params.Parts[0].String(), s.State.ID, s.Params.Parts[0], s.State.Balances[0]))
		require.Error(s.Adj.Deposit(s.Params.Parts[1].String(), s.State.ID, s.Params.Parts[1], s.State.Balances[1]))
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
			bal, err := s.Adj.BalanceOfID(s.Params.Parts[i].String())
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
			_, err := s.Adj.Withdraw(s.Parts[i].String(), s.State.ID, s.Parts[i])
			require.NoError(err)
		}

		// Token balance for parts must be the original value.
		for i := 0; i < 2; i++ {
			bal, err := s.Adj.BalanceOfID(s.Params.Parts[i].String()) // TODO: Fix parts
			require.Equal(s.State.Balances[i], bal)
			require.NoError(err)
		}
	})
}
