//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package asset_test

import (
	"github.com/perun-network/perun-fabric/channel/test"
	"github.com/perun-network/perun-fabric/wallet"
	req "github.com/stretchr/testify/require"
	"math/big"
	ptest "polycry.pt/poly-go/test"
	"testing"
)

func TestStubAsset(t *testing.T) {
	require := req.New(t)
	rng := ptest.Prng(ptest.NameStr("TestStubAsset"))
	initBal := 1000

	var sessions []*test.Session
	var accs []*wallet.Account
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		require.NoError(err)
		defer as.Close()
		sessions = append(sessions, as)
		accs = append(accs, wallet.NewRandomAccount(rng))
	}

	t.Run("Register", func(t *testing.T) {
		for i := uint(0); i < 2; i++ {
			// Register identity on address.
			err := sessions[i].Binding.RegisterAddress(accs[i].Address())
			req.NoError(t, err)

			// Check if initialized address got zero funds.
			bal, err := sessions[i].Binding.TokenBalance(accs[i].Address())
			req.NoError(t, err)
			req.Equal(t, big.NewInt(0), bal)
		}
	})

	t.Run("Double-Register", func(t *testing.T) {
		// Alice tries to re-register Bobs address for herself. Error expected.
		err := sessions[0].Binding.RegisterAddress(accs[1].Address())
		req.Error(t, err)
	})

	t.Run("Mint", func(t *testing.T) {
		for i := uint(0); i < 2; i++ {
			// Get current token balance. Expected to be zero.
			bal, err := sessions[i].Binding.TokenBalance(accs[i].Address())
			req.NoError(t, err)
			req.Equal(t, big.NewInt(0), bal)

			// Mint tokens.
			err = sessions[i].Binding.MintToken(accs[i].Address(), big.NewInt(int64(initBal)))
			req.NoError(t, err)

			// Get current token balance. Expected to be the minted amount.
			bal, err = sessions[i].Binding.TokenBalance(accs[i].Address())
			req.NoError(t, err)
			req.Equal(t, big.NewInt(int64(initBal)), bal)
		}
	})

	// The tests below can only succeed if minting above succeeded.
	// Otherwise, they are independent of another.

	t.Run("Transfer-Valid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)

		// Transfer 100.
		transfer := 100
		err = sessions[0].Binding.TokenToAddressTransfer(accs[0].Address(), accs[1].Address(), big.NewInt(int64(transfer)))
		req.NoError(t, err)

		// Check that balances changed as expected.
		bal, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		req.Equal(t, balAlice.Sub(balAlice, big.NewInt(int64(transfer))), bal)

		bal, err = sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)
		req.Equal(t, balBob.Add(balBob, big.NewInt(int64(transfer))), bal)
	})

	t.Run("Transfer-Zero-Invalid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)

		// Transfer zero. Expect error.
		transfer := 0
		err = sessions[0].Binding.TokenToAddressTransfer(accs[0].Address(), accs[1].Address(), big.NewInt(int64(transfer)))
		req.Error(t, err)

		// Ensure balances did not change.
		bal, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		req.Equal(t, balAlice, bal)
		bal, err = sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)
		req.Equal(t, balBob, bal)
	})

	t.Run("Transfer-Limit-Invalid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)

		// Transfer amount higher than Alice's funds. Expect error.
		transfer := big.NewInt(0)
		transfer.Add(balAlice, big.NewInt(1))
		err = sessions[0].Binding.TokenToAddressTransfer(accs[0].Address(), accs[1].Address(), transfer)
		req.Error(t, err)

		// Ensure balances did not change.
		bal, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		req.Equal(t, balAlice, bal)
		bal, err = sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)
		req.Equal(t, balBob, bal)
	})

	t.Run("Transfer-Fraudulent-Invalid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)

		// Bob tries to transfer 100 of Alice's token to his own address.
		transfer := 100
		err = sessions[1].Binding.TokenToAddressTransfer(accs[0].Address(), accs[1].Address(), big.NewInt(int64(transfer)))
		req.Error(t, err)

		// Ensure balances did not change.
		bal, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		req.Equal(t, balAlice, bal)
		bal, err = sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)
		req.Equal(t, balBob, bal)
	})

	t.Run("Burn", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)

		// Burn (all) tokens of both parties.
		err = sessions[0].Binding.BurnToken(accs[0].Address(), balAlice)
		req.NoError(t, err)
		err = sessions[1].Binding.BurnToken(accs[1].Address(), balBob)
		req.NoError(t, err)

		// Ensure balances are zero.
		bal, err := sessions[0].Binding.TokenBalance(accs[0].Address())
		req.NoError(t, err)
		req.Equal(t, big.NewInt(0), bal)
		bal, err = sessions[1].Binding.TokenBalance(accs[1].Address())
		req.NoError(t, err)
		req.Equal(t, big.NewInt(0), bal)
	})
}
