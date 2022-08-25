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
	"errors"
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type (
	// Ledger contains the read and write operations required by the Adjudicator.
	Ledger interface {
		StateLedger
		HoldingLedger
		Now() Timestamp
	}

	// StateLedger stores the channel's state.
	StateLedger interface {
		GetState(channel.ID) (*StateReg, error) //nolint:forbidigo
		PutState(*StateReg) error
	}

	// HoldingLedger stores the channel's holdings.
	HoldingLedger interface {
		GetHolding(channel.ID, wallet.Address) (*big.Int, error) //nolint:forbidigo
		PutHolding(channel.ID, wallet.Address, *big.Int) error
	}

	// NotFoundError should be returned by getters of Ledger implementations if
	// there's no entry under a given key.
	NotFoundError struct {
		Key  string
		Type string
	}
)

// IsNotFoundError returns true if given err is a NotFoundError.
func IsNotFoundError(err error) bool {
	notFoundErr := new(NotFoundError)
	return errors.As(err, &notFoundErr)
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("no entry for %q at %q", err.Type, err.Key)
}
