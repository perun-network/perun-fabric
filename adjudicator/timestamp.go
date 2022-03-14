// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"time"
)

type (
	// A Timestamp is a point in time that can be compared to others.
	Timestamp interface {
		Equal(Timestamp) bool
		After(Timestamp) bool
		Before(Timestamp) bool

		// Add adds the given duration to the Timestamp and returns it.
		// It is used to add challenge durations and thus should interpret the given
		// integer accordingly.
		Add(uint64) Timestamp

		// Clone returns a clone of the Timestamp.
		Clone() Timestamp
	}

	StdTimestamp time.Time
)

var newTimestamp = func() Timestamp { return StdTimestamp(time.Time{}) }

// NewTimestamp returns new Timestamp instances used during json unmarshaling of
// structs containing a Timestamp, e.g., StateRegs.
//
// It returns StdTimestamp instances by default.
func NewTimestamp() Timestamp { return newTimestamp() }

// SetNewTimestamp sets the Timestamp factory used during json unmarshaling of
// structs containing a Timestamp, e.g., StateRegs.
//
// It is set to return StdTimestamp instances by default.
func SetNewTimestamp(newts func() Timestamp) { newTimestamp = newts }

func StdNow() StdTimestamp {
	return (StdTimestamp)(time.Now())
}

func (t StdTimestamp) Time() time.Time { return (time.Time)(t) }

func asTime(t Timestamp) time.Time { return t.(StdTimestamp).Time() }

func (t StdTimestamp) Equal(other Timestamp) bool {
	return t.Time().Equal(asTime(other))
}

func (t StdTimestamp) After(other Timestamp) bool {
	return t.Time().After(asTime(other))
}

func (t StdTimestamp) Before(other Timestamp) bool {
	return t.Time().Before(asTime(other))
}

func (t StdTimestamp) Add(d uint64) Timestamp {
	ts := t.Time().Add(time.Duration(d) * time.Second)
	return (StdTimestamp)(ts)
}

func (t StdTimestamp) Clone() Timestamp {
	return t
}

func (t StdTimestamp) MarshalJSON() ([]byte, error) {
	return (time.Time)(t).MarshalJSON()
}

func (t *StdTimestamp) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}
