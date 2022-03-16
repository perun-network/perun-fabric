// SPDX-License-Identifier: Apache-2.0

package test

import adj "github.com/perun-network/perun-fabric/adjudicator"

// TestLedger is a MemLedger with a controllable clock.
//
// The timestamp returned by Now stays constant until it is advanced with
// AdvanceNow.
// The initial value of now is set at creation to MemLedger.Now().
type TestLedger struct {
	*adj.MemLedger
	now adj.Timestamp
}

func NewTestLedger() *TestLedger {
	ml := adj.NewMemLedger()
	return &TestLedger{
		MemLedger: ml,
		now:       ml.Now(),
	}
}

func (l *TestLedger) Now() adj.Timestamp {
	return l.now
}

func (l *TestLedger) AdvanceNow(duration uint64) {
	l.now = l.now.Add(duration)
}
