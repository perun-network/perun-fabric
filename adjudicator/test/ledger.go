// SPDX-License-Identifier: Apache-2.0

package test

import adj "github.com/perun-network/perun-fabric/adjudicator"

// TestLedger is a MemLedger with a controllable clock.
// The timestamp returned by Now stays constant until it is advanced with AdvanceNow.
// The initial value of now is set at creation to MemLedger.Now().
type TestLedger struct { //nolint:revive
	*adj.MemLedger
	now adj.Timestamp
}

// NewTestLedger creates a test ledger with a controllable clock.
func NewTestLedger() *TestLedger {
	ml := adj.NewMemLedger()
	return &TestLedger{
		MemLedger: ml,
		now:       ml.Now(),
	}
}

// Now returns the current time on the TestLedger.
func (l *TestLedger) Now() adj.Timestamp {
	return l.now
}

// AdvanceNow adds the given duration onto the current time of the TestLedger.
func (l *TestLedger) AdvanceNow(duration uint64) {
	l.now = l.now.Add(duration)
}
