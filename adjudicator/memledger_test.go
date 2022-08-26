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
