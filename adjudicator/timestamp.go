// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"time"
)

type (
	// A Timestamp is a point in time that can be compared to others.
	Timestamp time.Time
)

var newTimestamp = func() Timestamp { return Timestamp(time.Time{}) }

// NewTimestamp returns new Timestamp instances used during json unmarshaling of
// structs containing a Timestamp, e.g., StateRegs.
//
// It returns Timestamp instances by default.
func NewTimestamp() Timestamp { return newTimestamp() }

// SetNewTimestamp sets the Timestamp factory used during json unmarshaling of
// structs containing a Timestamp, e.g., StateRegs.
//
// It is set to return Timestamp instances by default.
func SetNewTimestamp(newts func() Timestamp) { newTimestamp = newts }

func StdNow() Timestamp {
	return (Timestamp)(time.Now())
}

func (t Timestamp) Time() time.Time { return (time.Time)(t) }

func asTime(t Timestamp) time.Time { return t.Time() }

func (t Timestamp) Equal(other Timestamp) bool {
	return t.Time().Equal(asTime(other))
}

func (t Timestamp) After(other Timestamp) bool {
	return t.Time().After(asTime(other))
}

func (t Timestamp) Before(other Timestamp) bool {
	return t.Time().Before(asTime(other))
}

func (t Timestamp) Add(d uint64) Timestamp {
	ts := t.Time().Add(time.Duration(d) * time.Second)
	return (Timestamp)(ts)
}

func (t Timestamp) Clone() Timestamp {
	return t
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return (time.Time)(t).MarshalJSON()
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}
