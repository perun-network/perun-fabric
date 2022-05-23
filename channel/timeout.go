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
	t       time.Time
	polling time.Duration
}

func makeTimeout(
	t time.Time,
	polling time.Duration,
) *Timeout {
	return &Timeout{
		t:       t,
		polling: polling,
	}
}

// IsElapsed should return whether the timeout has elapsed at the time of
// the call of this method.
func (t *Timeout) IsElapsed(ctx context.Context) bool {
	return true

	// TODO: There is no block time in fabric (yet) (see links below)
	// https://jira.hyperledger.org/browse/FAB-15584
	// https://stackoverflow.com/questions/59135319/does-the-block-contain-the-generation-time-of-the-block

	// resp, err := t.c.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})

	// if err != nil {
	//	log.Printf("Warning: Error getting latest block: %v\n", err)
	//	return false
	// }
	// return resp.Block.Header.Time.After(t.t)
}

// Wait waits for the timeout to elapse. If the context is canceled, Wait
// should return immediately with the context's error.
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
