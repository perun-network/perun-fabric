// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type (
	// A Timestamp is a point in time that can be compared to others.
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

var newTimestamp = func() Timestamp { return (*StdTimestamp)(new(time.Time)) }

func NewTimestamp() Timestamp { return newTimestamp() }

// SetTimestampConstructor can be used in any init() function to set the
// constructor returning new Timestamp instances of the concrete type used in
// the current runtime context. It is used during JSON unmarshaling. The set
// function can be accessed via NewTimestamp.
//
// By default, it returns instances of (*StdTimestamp)(new(time.Time)).
func SetTimestampConstructor(newts func() Timestamp) {
	newTimestamp = newts
}

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
