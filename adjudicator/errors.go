// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"
)

var ErrUnknownChannel = errors.New("unknown channel")

type (
	ValidationError struct{ error }

	ChallengeTimeoutError struct {
		Timeout RegTimestamp
		Now     RegTimestamp
	}

	VersionError struct {
		Registered uint64
		Tried      uint64
	}

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
