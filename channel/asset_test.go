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

package channel_test

import (
	"github.com/perun-network/perun-fabric/channel/test"
	requ "github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestStubAsset(t *testing.T) {
	require := requ.New(t)

	var sessions []*test.Session
	for i := uint(1); i <= 2; i++ {
		as, err := test.NewTestSession(test.OrgNum(i), test.AdjudicatorName)
		require.NoError(err)
		defer as.Close()
		sessions = append(sessions, as)
	}

	t.Run("Mint-Valid", func(t *testing.T) {
		mintingBal := big.NewInt(100)

		// Get current token balance.
		bal, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)

		// Mint tokens.
		err = sessions[0].Binding.MintToken(mintingBal)
		requ.NoError(t, err)

		// Get current token balance. Expected to be the minted amount.
		newBal, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)

		// Calculate expected balance.
		expectedBal := big.NewInt(0)
		expectedBal.Add(bal, mintingBal)

		requ.Equal(t, expectedBal, newBal)
	})

	t.Run("Mint-Rejected", func(t *testing.T) {
		// Note that session 1 is not central banker, hence not allowed to mint tokens.
		mintingBal := big.NewInt(100)

		// Get current token balance.
		bal, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// Mint tokens.
		err = sessions[1].Binding.MintToken(mintingBal)
		requ.Error(t, err)

		// Get current token balance. Expected to be the minted amount.
		newBal, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// We expect no change in balance.
		requ.Equal(t, bal, newBal)
	})

	t.Run("Transfer-Valid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// Transfer 100.
		transfer := big.NewInt(100)
		err = sessions[0].Binding.TokenTransfer(sessions[1].ClientFabricID, transfer)
		requ.NoError(t, err)

		// Check that balances changed as expected.
		bal, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balAlice.Sub(balAlice, transfer), bal)

		bal, err = sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balBob.Add(balBob, transfer), bal)
	})

	t.Run("Transfer-Negative-Invalid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// Transfer zero. Expect error.
		transfer := big.NewInt(-1)
		err = sessions[0].Binding.TokenTransfer(sessions[1].ClientFabricID, transfer)
		requ.Error(t, err)

		// Ensure balances did not change.
		bal, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balAlice, bal)
		bal, err = sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balBob, bal)
	})

	t.Run("Transfer-Limit-Invalid", func(t *testing.T) {
		// Get current balances.
		balAlice, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		balBob, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// Transfer amount higher than Alice's funds. Expect error.
		transfer := big.NewInt(0)
		transfer.Add(balAlice, big.NewInt(1))
		err = sessions[0].Binding.TokenTransfer(sessions[1].ClientFabricID, transfer)
		requ.Error(t, err)

		// Ensure balances did not change.
		bal, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balAlice, bal)
		bal, err = sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, balBob, bal)
	})

	t.Run("Burn", func(t *testing.T) {
		// Get current balances.
		initBalAlice, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		initBalBob, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)

		// Burn tokens.
		burnAmount := big.NewInt(5)
		err = sessions[0].Binding.BurnToken(burnAmount)
		requ.NoError(t, err)

		expBalAlice := big.NewInt(0)
		expBalAlice.Sub(initBalAlice, burnAmount)

		// Ensure balances changed accordingly.
		newBalAlice, err := sessions[0].Binding.TokenBalance(sessions[0].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, expBalAlice, newBalAlice)
		newBalBob, err := sessions[1].Binding.TokenBalance(sessions[1].ClientFabricID)
		requ.NoError(t, err)
		requ.Equal(t, initBalBob, newBalBob)
	})
}
