// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"math/big"
	"math/rand"
	"testing"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"

	"polycry.pt/poly-go/test"
)

func TestAssetHolder(t *testing.T) {
	rng := test.Prng(t)

	t.Run("MultiDepositWithdraw", func(t *testing.T) {
		require := require.New(t)
		ah, id, _addrs, bals, _ := ahSetup(rng, 2)
		addr := _addrs[0]

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
		ah, id, addrs, bals, params := ahSetup(rng, n)

		totalh, err := ah.TotalHolding(params)
		require.NoError(err)
		require.Zero(totalh.Sign())

		total := new(big.Int)
		for i, addr := range addrs {
			require.NoError(ah.Deposit(id, addr, bals[i]))
			total.Add(total, bals[i])
			totalh, err = ah.TotalHolding(params)
			require.NoError(err)
			require.Zero(total.Cmp(totalh))
		}
	})
}

func ahSetup(rng *rand.Rand, n int) (*adj.AssetHolder, channel.ID, []wallet.Address, []channel.Bal, *channel.Params) {
	params := chtest.NewRandomParams(rng, chtest.WithNumParts(n))
	return memAssetHolder(),
		params.ID(),
		params.Parts,
		chtest.NewRandomBals(rng, n),
		params
}

func memAssetHolder() *adj.AssetHolder {
	return adj.NewAssetHolder(adj.NewMemLedger())
}
