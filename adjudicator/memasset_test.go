package adjudicator_test

import (
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/wallet"
	"github.com/stretchr/testify/require"
	"math/big"
	chtest "perun.network/go-perun/channel/test"
	"polycry.pt/poly-go/test"
	"testing"
)

func TestMemAsset(t *testing.T) {
	rng := test.Prng(t)

	t.Run("Mint", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		acc := wallet.NewRandomAccount(rng)
		addr := acc.Address()

		// Mint (incremental).
		require.NoError(ma.Mint("", addr, big.NewInt(100)))
		require.NoError(ma.Mint("", addr, big.NewInt(50)))

		// Check balance of address.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOfAddress(addr)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Burn", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		acc := wallet.NewRandomAccount(rng)
		addr := acc.Address()

		// Mint.
		require.NoError(ma.Mint("", addr, big.NewInt(150)))

		// Burn.
		require.NoError(ma.Burn("", addr, big.NewInt(100)))

		// Check balance of address.
		expectedBal := big.NewInt(50)
		bal, err := ma.BalanceOfAddress(addr)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-to-address", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address()
		addrTwo := wallet.NewRandomAccount(rng).Address()

		// Mint.
		require.NoError(ma.Mint("", addrOne, big.NewInt(150)))
		require.NoError(ma.Mint("", addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.NoError(ma.AddressToAddressTransfer("", addrOne, addrTwo, big.NewInt(50)))
		require.NoError(ma.AddressToAddressTransfer("", addrOne, addrTwo, big.NewInt(50)))
		require.NoError(ma.AddressToAddressTransfer("", addrTwo, addrOne, big.NewInt(25)))

		// Check balance of address one.
		expectedBal := big.NewInt(75)
		bal, err := ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(225)
		bal, err = ma.BalanceOfAddress(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-zero", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address()
		addrTwo := wallet.NewRandomAccount(rng).Address()

		// Mint.
		require.NoError(ma.Mint("", addrOne, big.NewInt(150)))
		require.NoError(ma.Mint("", addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.Error(ma.AddressToAddressTransfer("", addrOne, addrTwo, big.NewInt(0)))

		// Check balance of address one.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(150)
		bal, err = ma.BalanceOfAddress(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-not-enough-funds", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		addrOne := wallet.NewRandomAccount(rng).Address()
		addrTwo := wallet.NewRandomAccount(rng).Address()

		// Mint.
		require.NoError(ma.Mint("", addrOne, big.NewInt(50)))
		require.NoError(ma.Mint("", addrTwo, big.NewInt(150)))

		// Transfer (incremental).
		require.NoError(ma.AddressToAddressTransfer("", addrOne, addrTwo, big.NewInt(50)))
		require.Error(ma.AddressToAddressTransfer("", addrOne, addrTwo, big.NewInt(100)))

		// Check balance of address one.
		expectedBal := big.NewInt(0)
		bal, err := ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of address two.
		expectedBal = big.NewInt(200)
		bal, err = ma.BalanceOfAddress(addrTwo)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-to-channel", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		channelID := chtest.NewRandomChannelID(rng)
		addrOne := wallet.NewRandomAccount(rng).Address()

		// Mint.
		require.NoError(ma.Mint("", addrOne, big.NewInt(150)))

		// Transfer (incremental).
		require.NoError(ma.AddressToChannelTransfer("", addrOne, channelID, big.NewInt(50)))
		require.NoError(ma.AddressToChannelTransfer("", addrOne, channelID, big.NewInt(50)))

		// Check balance of address.
		expectedBal := big.NewInt(50)
		bal, err := ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of channel.
		expectedBal = big.NewInt(100)
		bal, err = ma.BalanceOfChannel(channelID)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-channel-to-address", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		channelID := chtest.NewRandomChannelID(rng)
		addrOne := wallet.NewRandomAccount(rng).Address()

		// Mint and fund channel.
		require.NoError(ma.Mint("", addrOne, big.NewInt(150)))
		require.NoError(ma.AddressToChannelTransfer("", addrOne, channelID, big.NewInt(150)))

		// Check balance of channel.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOfChannel(channelID)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Transfer (incremental).
		require.NoError(ma.ChannelToAddressTransfer(channelID, addrOne, big.NewInt(50)))
		require.NoError(ma.ChannelToAddressTransfer(channelID, addrOne, big.NewInt(50)))

		// Check balance of address.
		expectedBal = big.NewInt(100)
		bal, err = ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of channel.
		expectedBal = big.NewInt(50)
		bal, err = ma.BalanceOfChannel(channelID)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})

	t.Run("Transfer-channel-to-address-not-enough-funds", func(t *testing.T) {
		require := require.New(t)
		ma := adj.NewMemAsset()
		channelID := chtest.NewRandomChannelID(rng)
		addrOne := wallet.NewRandomAccount(rng).Address()

		// Mint and fund channel.
		require.NoError(ma.Mint("", addrOne, big.NewInt(150)))
		require.NoError(ma.AddressToChannelTransfer("", addrOne, channelID, big.NewInt(150)))

		// Check balance of channel.
		expectedBal := big.NewInt(150)
		bal, err := ma.BalanceOfChannel(channelID)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Transfer (incremental).
		require.NoError(ma.ChannelToAddressTransfer(channelID, addrOne, big.NewInt(100)))
		require.Error(ma.ChannelToAddressTransfer(channelID, addrOne, big.NewInt(100)))

		// Check balance of address.
		expectedBal = big.NewInt(100)
		bal, err = ma.BalanceOfAddress(addrOne)
		require.NoError(err)
		require.Equal(expectedBal, bal)

		// Check balance of channel.
		expectedBal = big.NewInt(50)
		bal, err = ma.BalanceOfChannel(channelID)
		require.NoError(err)
		require.Equal(expectedBal, bal)
	})
}
