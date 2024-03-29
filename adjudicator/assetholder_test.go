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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"

	_ "github.com/perun-network/perun-fabric" // init backend
	adj "github.com/perun-network/perun-fabric/adjudicator"
)

func TestAssetHolder(t *testing.T) {
	rng := test.Prng(t)

	t.Run("SetHolding", func(t *testing.T) {
		require := require.New(t)
		ah, id, addrs, bals := ahSetup(rng, 1)
		addr, bal := addrs[0], bals[0]

		hzero, err := ah.Holding(id, addr)
		require.NoError(err)
		require.Zero(hzero.Sign())

		require.NoError(ah.SetHolding(id, addr, bal))
		h, err := ah.Holding(id, addr)
		require.NoError(err)
		require.Zero(h.Cmp(bal))
	})

	t.Run("MultiDepositWithdraw", func(t *testing.T) {
		require := require.New(t)
		ah, id, addrs, bals := ahSetup(rng, 2)
		addr := addrs[0]

		hzero, err := ah.Holding(id, addr)
		require.NoError(err)
		require.Zero(hzero.Sign())

		// 1st deposit
		require.NoError(ah.Deposit(id, addr, bals[0]))
		h, err := ah.Holding(id, addr)
		require.NoError(err)
		require.Zero(h.Cmp(bals[0]))

		// 2nd deposit
		require.NoError(ah.Deposit(id, addr, bals[1]))
		h1, err := ah.Holding(id, addr)
		require.NoError(err)
		total := new(big.Int).Add(bals[0], bals[1])
		require.Zero(h1.Cmp(total))

		wbal, err := ah.Withdraw(id, addr)
		require.NoError(err)
		require.Zero(wbal.Cmp(total))

		hfinal, err := ah.Holding(id, addr)
		require.NoError(err)
		require.Zero(hfinal.Sign())

		// 2nd withdrawal should be 0
		wbal1, err := ah.Withdraw(id, addr)
		require.NoError(err)
		require.Zero(wbal1.Sign())
	})

	t.Run("TotalHolding", func(t *testing.T) {
		const n = 3
		require := require.New(t)
		ah, id, addrs, bals := ahSetup(rng, n)

		totalh, err := ah.TotalHolding(id, addrs)
		require.NoError(err)
		require.Zero(totalh.Sign())

		total := new(big.Int)
		for i, addr := range addrs {
			require.NoError(ah.Deposit(id, addr, bals[i]))
			total.Add(total, bals[i])
			totalh, err = ah.TotalHolding(id, addrs)
			require.NoError(err)
			require.Zero(total.Cmp(totalh))
		}
	})
}

func ahSetup(rng *rand.Rand, n int) (*adj.AssetHolder, channel.ID, []wallet.Address, []channel.Bal) {
	return memAssetHolder(),
		chtest.NewRandomChannelID(rng),
		wtest.NewRandomAddresses(rng, n),
		chtest.NewRandomBals(rng, n)
}

func memAssetHolder() *adj.AssetHolder {
	return adj.NewAssetHolder(adj.NewMemLedger())
}
