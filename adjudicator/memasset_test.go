package adjudicator_test

import (
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/wallet"
	"github.com/stretchr/testify/require"
	"math/big"
	"polycry.pt/poly-go/test"
	"testing"
)

func TestMemAsset(t *testing.T) {
	rng := test.Prng(t)

	t.Run("Mint", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		acc := wallet.NewRandomAccount(rng)
		addr := acc.Address().String()

		// Mint (incremental).
		require.NoError(ma.Mint(addr, big.NewInt(100)))
		require.NoError(ma.Mint(addr, big.NewInt(50)))

		// Check balance of address.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOf(addr)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Burn", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		acc := wallet.NewRandomAccount(rng)
		addr := acc.Address().String()

		// Mint.
		require.NoError(ma.Mint(addr, big.NewInt(150)))

		// Burn.
		require.NoError(ma.Burn(addr, big.NewInt(100)))

		// Check balance of address.
		expectedBal := big.NewInt(50)
		bal, err := ma.BalanceOf(addr)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-to-address", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address().String()
		addrTwo := wallet.NewRandomAccount(rng).Address().String()

		// Mint.
		require.NoError(ma.Mint(addrOne, big.NewInt(150)))
		require.NoError(ma.Mint(addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.NoError(ma.Transfer(addrOne, addrTwo, big.NewInt(50)))
		require.NoError(ma.Transfer(addrOne, addrTwo, big.NewInt(50)))
		require.NoError(ma.Transfer(addrTwo, addrOne, big.NewInt(25)))

		// Check balance of address one.
		expectedBal := big.NewInt(75)
		bal, err := ma.BalanceOf(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(225)
		bal, err = ma.BalanceOf(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-zero", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address().String()
		addrTwo := wallet.NewRandomAccount(rng).Address().String()

		// Mint.
		require.NoError(ma.Mint(addrOne, big.NewInt(150)))
		require.NoError(ma.Mint(addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.Error(ma.Transfer(addrOne, addrTwo, big.NewInt(0)))

		// Check balance of address one.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOf(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(150)
		bal, err = ma.BalanceOf(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-not-enough-funds", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address().String()
		addrTwo := wallet.NewRandomAccount(rng).Address().String()

		// Mint.
		require.NoError(ma.Mint(addrOne, big.NewInt(50)))
		require.NoError(ma.Mint(addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.NoError(ma.Transfer(addrOne, addrTwo, big.NewInt(50)))
		require.Error(ma.Transfer(addrOne, addrTwo, big.NewInt(100)))

		// Check balance of address one.
		expectedBal := big.NewInt(0)
		bal, err := ma.BalanceOf(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(200)
		bal, err = ma.BalanceOf(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})
}
