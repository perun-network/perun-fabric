// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"
)

// ErrUnknownChannel indicates that no information for a channel id could be found by the chaincode.
var ErrUnknownChannel = errors.New("unknown channel")

type (
	// ValidationError indicates that the given arguments could not be validated successfully.
	ValidationError struct{ error }

	// ChallengeTimeoutError indicates that the challenge timeout passed.
	ChallengeTimeoutError struct {
		Timeout Timestamp
		Now     Timestamp
	}

	// VersionError indicates that the chaincode holds a newer version of the proposed channel state.
	VersionError struct {
		Registered uint64
		Tried      uint64
	}

	// UnderfundedError indicates that the sum of the proposed balances are higher than the actual funding.
	UnderfundedError struct {
		Version uint64
		Total   *big.Int
		Funded  *big.Int
	}
)

func (ve ValidationError) Unwrap() error {
	return ve.error
}

func (te ChallengeTimeoutError) Error() string {
	return fmt.Sprintf("challenge period ended (timeout: %v, now: %v)", te.Timeout, te.Now)
}

func (ve VersionError) Error() string {
	return fmt.Sprintf("version too low (registered: %d, tried: %d)", ve.Registered, ve.Tried)
}

func (ue UnderfundedError) Error() string {
	return fmt.Sprintf("channel underfunded (%v < %v, version %d)", ue.Funded, ue.Total, ue.Version)
}

// IsAdjudicatorError returns true if the given error is one of the following:
// ValidationError, ChallengeTimeoutError, VersionError, UnderfundedError.
func IsAdjudicatorError(err error) bool {
	if err == nil {
		return false
	}

	adjErrors := []interface{}{
		new(ValidationError),
		new(ChallengeTimeoutError),
		new(VersionError),
		new(UnderfundedError),
	}
	for _, aerr := range adjErrors {
		if errors.As(err, aerr) {
			return true
		}
	}

	return false
}
