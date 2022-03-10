// SPDX-License-Identifier: Apache-2.0

package adjudicator

import "time"

type (
	// A Timestamp is a point in time that can be compared to others.
	Timestamp interface {
		Equal(Timestamp) bool
		After(Timestamp) bool
		Before(Timestamp) bool

		// Clone returns a clone of the Timestamp.
		Clone() Timestamp
	}

	StdTimestamp time.Time
)

func StdNow() *StdTimestamp {
	now := time.Now()
	return (*StdTimestamp)(&now)
}

func (t StdTimestamp) Time() time.Time { return (time.Time)(t) }

func asTime(t Timestamp) time.Time { return t.(*StdTimestamp).Time() }

func (t StdTimestamp) Equal(other Timestamp) bool {
	return t.Time().Equal(asTime(other))
}

func (t StdTimestamp) After(other Timestamp) bool {
	return t.Time().After(asTime(other))
}

func (t StdTimestamp) Before(other Timestamp) bool {
	return t.Time().Before(asTime(other))
}
