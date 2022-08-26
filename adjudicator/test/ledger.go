// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
