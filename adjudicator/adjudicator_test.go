// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"math/big"
	"testing"

	"github.com/go-test/deep"
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
		s := adjtest.NewSetup(test.Prng(t))

		for i := 0; i < 2; i++ {
			h, err := s.Adj.Holding(s.State.ID, s.Params.Parts[i])
			require.NoError(err)
			require.Zero(h.Sign())
		}

		th, err := s.Adj.TotalHolding(s.State.ID, s.Params.Parts)
		require.NoError(err)
		require.Zero(th.Sign())

		// Deposit twice each to test additivity
		for i := 0; i < 2; i++ {
			require.NoError(s.Adj.Deposit(s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
			require.NoError(s.Adj.Deposit(s.State.ID, s.Params.Parts[i], s.State.Balances[i]))
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
		require.Zero(deep.Equal(sr, adjsr))
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
		require.Zero(deep.Equal(sr, adjsr))
	})

	t.Run("Register-refute", func(t *testing.T) {
		require := require.New(t)
		s := adjtest.NewSetup(test.Prng(t), adjtest.Funded)

		ch0 := s.SignedChannel()
		require.NoError(s.Adj.Register(ch0))

		// increment state's version
		s.State.Version = 5
		// forward clock close to dispute timeout
		s.Ledger.AdvanceNow(s.Params.ChallengeDuration)

		ch1 := s.SignedChannel()
		sr1 := s.StateReg()
		require.NoError(s.Adj.Register(ch1))

		adjsr1, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.Zero(deep.Equal(sr1, adjsr1))
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
		require.Zero(deep.Equal(sr0, adjsr0))
	})
}
