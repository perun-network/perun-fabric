// SPDX-License-Identifier: Apache-2.0

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
