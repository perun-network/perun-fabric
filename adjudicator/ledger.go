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
	Timestamp interface {
		Equal(Timestamp) bool
		After(Timestamp) bool
		Before(Timestamp) bool

		// Clone returns a clone of the Timestamp.
		Clone() Timestamp
	}

	// Ledger contains the read and write operations required by the Adjudicator.
	Ledger interface {
		StateLedger
		HoldingLedger

		Now() Timestamp
	}

	StateLedger interface {
		GetState(channel.ID) (*StateReg, error)
		PutState(*StateReg) error
	}

	StateReg struct {
		*channel.State
		Timeout Timestamp
	}

	HoldingLedger interface {
		GetHolding(channel.ID, wallet.Address) (*big.Int, error)
		PutHolding(channel.ID, wallet.Address, *big.Int) error
	}

	// NotFoundError should be returned by getters of Ledger implementations if
	// there's no entry under a given key.
	NotFoundError struct {
		Key  string
		Type string
	}
)

func IsNotFoundError(err error) bool {
	notFoundErr := new(NotFoundError)
	return errors.As(err, &notFoundErr)
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("no entry for %q at %q", err.Type, err.Key)
}

func (s *StateReg) Clone() *StateReg {
	return &StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone(),
	}
}