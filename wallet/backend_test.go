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
