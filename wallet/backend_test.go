// SPDX-License-Identifier: Apache-2.0

package wallet_test

import (
	"math/big"
	"testing"

	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"

	"github.com/perun-network/perun-fabric/wallet"
	"github.com/stretchr/testify/require"
)

func TestAccountWalletBackend(t *testing.T) {
	rng := test.Prng(t)
	acc := wallet.NewRandomAccount(rng)
	addr := acc.Address().(*wallet.Address)
	addrUnk := wallet.NewRandomAddress(rng)
	addrUnkBytes, err := addrUnk.MarshalBinary()
	require.NoError(t, err)

	setup := &wtest.Setup{
		Backend:         wallet.Backend{},
		Wallet:          wallet.NewWallet(acc),
		AddressInWallet: addr,
		ZeroAddress: &wallet.Address{
			Curve: addr.Curve,
			X:     new(big.Int),
			Y:     new(big.Int),
		},
		DataToSign:        addr.X.Bytes(),
		AddressMarshalled: addrUnkBytes,
	}

	wtest.TestAccountWithWalletAndBackend(t, setup)
}
