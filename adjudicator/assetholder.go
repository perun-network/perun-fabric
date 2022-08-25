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
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// AssetHolder tracks deposits and withdrawals of channel participants over a
// HoldingLedger.
type AssetHolder struct {
	ledger HoldingLedger
}

// NewAssetHolder returns a new AssetHolder operating on the given ledger.
func NewAssetHolder(ledger HoldingLedger) *AssetHolder {
	return &AssetHolder{ledger: ledger}
}

// Deposit registers a deposit for channel `id` and participant `part` of amount
// `amount`, possibly adding to an already existent deposit.
//
// Deposit throws an error if `amount` is negative.
//
// Ledger access errors are propagated.
func (a *AssetHolder) Deposit(id channel.ID, part wallet.Address, amount *big.Int) error {
	if amount.Sign() == -1 {
		return fmt.Errorf("negative amount")
	}

	holding := new(big.Int)
	if current, err := a.ledger.GetHolding(id, part); err == nil {
		holding.Set(current)
	} else if !IsNotFoundError(err) {
		return fmt.Errorf("querying ledger holding: %w", err)
	}
	holding.Add(holding, amount)

	if err := a.ledger.PutHolding(id, part, holding); err != nil {
		return fmt.Errorf("putting ledger holding: %w", err)
	}

	return nil
}

// Holding returns the holdings of participant `part` in the channel of id `id`.
func (a *AssetHolder) Holding(id channel.ID, part wallet.Address) (*big.Int, error) {
	holding := new(big.Int)
	if current, err := a.ledger.GetHolding(id, part); err == nil {
		holding.Set(current)
	} else if !IsNotFoundError(err) {
		return nil, fmt.Errorf("querying ledger holding: %w", err)
	}
	return holding, nil
}

// TotalHolding returns the total amount deposited into the channel specified by
// `params`.
func (a *AssetHolder) TotalHolding(id channel.ID, parts []wallet.Address) (*big.Int, error) {
	total := new(big.Int)
	for _, part := range parts {
		holding, err := a.Holding(id, part)
		if err != nil {
			return nil, err
		}
		total.Add(total, holding)
	}
	return total, nil
}

// SetHolding sets the holding of part in channel id to holding.
//
// Panics if `holding` is negative.
func (a *AssetHolder) SetHolding(id channel.ID, part wallet.Address, holding *big.Int) error {
	if holding.Sign() == -1 {
		return fmt.Errorf("negative amount")
	}
	return a.ledger.PutHolding(id, part, holding)
}

// Withdraw resets the holdings of participant `part` in the channel of id `id`
// to zero and returns the holdings before the reset.
func (a *AssetHolder) Withdraw(id channel.ID, part wallet.Address) (*big.Int, error) {
	holding, err := a.Holding(id, part)
	if err != nil {
		return nil, err
	}
	if err = a.ledger.PutHolding(id, part, new(big.Int)); err != nil {
		return nil, fmt.Errorf("zeroing ledger holding: %w", err)
	}
	return holding, nil
}
