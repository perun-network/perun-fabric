//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package channel

import (
	"context"
	"time"
)

// Timeout represents a timeout that is bound to block time.
type Timeout struct {
	timeout time.Time
	polling time.Duration
}

// MakeTimeout generates a timeout based on the given duration and the current time.
func MakeTimeout(duration time.Duration, polling time.Duration) *Timeout {
	t := time.Now().Add(duration)

	return &Timeout{
		timeout: t,       // timeout is the time representing the timeout
		polling: polling, // polling is used to periodically check if the timeout is concluded
	}
}

// IsElapsed should return whether the timeout has concluded at the time of the call of this method.
func (t *Timeout) IsElapsed(ctx context.Context) bool {
	current := time.Now()           // Fabric does not have a block time.
	return current.After(t.timeout) // Instead, use the system time to compare against.
}

// Wait waits for the timeout to elapse.
// If the context is canceled, Wait returns immediately with the context's error.
func (t *Timeout) Wait(ctx context.Context) error {
	for !t.IsElapsed(ctx) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(t.polling):
		}
	}
	return nil
}
