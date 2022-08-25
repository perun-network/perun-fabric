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

package adjudicator

import (
	"math/big"
)

// AccountID represents the ID of a client.
// It is used for minting, burning, receiving and sending funds.
// Ensure it is unique for every client interacting with Asset and no impersonation is possible.
type AccountID string

// Asset is a basic interface for creating tokens with.
type Asset interface {
	// Mint creates the desired amount of token for the given id.
	// Note that id must be authenticated first.
	Mint(id AccountID, amount *big.Int) error

	// Burn removes the desired amount of token from the given id.
	// Note that id must be authenticated first.
	Burn(id AccountID, amount *big.Int) error

	// Transfer sends the desired amount of tokens from sender to receiver.
	// Note that sender must be authenticated first.
	Transfer(sender AccountID, receiver AccountID, amount *big.Int) error

	// BalanceOf returns the amount of tokens the given id holds.
	BalanceOf(id AccountID) (*big.Int, error)
}
