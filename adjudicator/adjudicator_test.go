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

		ch := s.SignedChannel()
		require.NoError(s.Adj.Register(ch))

		reg, err := s.Adj.StateReg(s.State.ID)
		require.NoError(err)
		require.Zero(deep.Equal(&adj.StateReg{
			State:   *s.State,
			Timeout: s.Ledger.Now(),
		}, reg))
	})
}
