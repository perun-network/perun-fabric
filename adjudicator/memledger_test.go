// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"testing"
	"time"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chtest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"
)

func TestStdTime(t *testing.T) {
	t0, t1 := adj.StdNow(), adj.StdNow()

	t.Run("Equal", func(t *testing.T) {
		assert.True(t, t0.Equal(t0))
		assert.True(t, t1.Equal(t1))
		assert.False(t, t0.Equal(t1))
	})

	t.Run("After", func(t *testing.T) {
		assert.False(t, t0.After(t1))
		assert.False(t, t0.After(t0))
		assert.True(t, t1.After(t0))
	})

	t.Run("Before", func(t *testing.T) {
		assert.True(t, t0.Before(t1))
		assert.False(t, t0.Before(t0))
		assert.False(t, t1.Before(t0))
	})

	t.Run("Clone", func(t *testing.T) {
		t0c := t0.Clone().(*adj.StdTimestamp)
		(*t0c) = (adj.StdTimestamp)(t0c.Time().Add(time.Hour))
		assert.False(t, t0c.Equal(t0))
	})
}

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
