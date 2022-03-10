// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"testing"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"

	"github.com/stretchr/testify/require"
	chtest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"
)

func TestMemLedger(t *testing.T) {
	rng := test.Prng(t)

	t.Run("State", func(t *testing.T) {
		var (
			require = require.New(t)
			ml      = adj.NewMemLedger()
			sr      = adjtest.RandomStateReg(rng)
		)

		sget, err := ml.GetState(sr.ID)
		require.Nil(sget)
		require.True(adj.IsNotFoundError(err))

		require.NoError(ml.PutState(sr))

		sget, err = ml.GetState(sr.ID)
		require.Equal(sr, sget)
		require.NoError(err)
	})

	t.Run("Holding", func(t *testing.T) {
		var (
			require = require.New(t)
			ml      = adj.NewMemLedger()
			id      = chtest.NewRandomChannelID(rng)
			addr    = wtest.NewRandomAddress(rng)
			bal     = chtest.NewRandomBal(rng)
		)

		hget, err := ml.GetHolding(id, addr)
		require.Nil(hget)
		require.True(adj.IsNotFoundError(err))

		require.NoError(ml.PutHolding(id, addr, bal))

		hget, err = ml.GetHolding(id, addr)
		require.Equal(bal, hget)
		require.NoError(err)
	})
}
