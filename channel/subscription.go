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
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	fabclient "github.com/perun-network/perun-fabric/client"
	"sync"
	"time"

	"perun.network/go-perun/channel"
)

// EventSubscription provides methods for consuming channel events.
type EventSubscription struct {
	adjudicator *Adjudicator    // adjudicator is the referenced adjudicator instance.
	channelID   channel.ID      // channelID is the channel identifier
	prevState   adj.StateReg    // prevState is the previous channel state
	timeout     *Timeout        // timeout is the current Event timeout
	concluded   bool            // concluded indicates if a concluded event was created
	registered  bool            // registered indicates if any state has been registered on-chain
	err         chan error      // err forwards errors during event parsing
	once        sync.Once       // once used to close channels
	ctx         context.Context // ctx is the context to establish the subscription
}

func NewEventSubscription(ctx context.Context, a *Adjudicator, ch channel.ID) (*EventSubscription, error) {
	return &EventSubscription{
		adjudicator: a,
		channelID:   ch,
		err:         make(chan error, 1),
		prevState:   adj.StateReg{},
		timeout:     nil,
		ctx:         ctx,
	}, nil
}

// Next returns the most recent or next future event. If the subscription is
// closed or any other error occurs, it returns nil.
func (s *EventSubscription) Next() channel.AdjudicatorEvent {
	for {
		// Get the on chain state.
		d, err := s.getState()
		if err != nil {
			s.err <- err
			return nil
		}

		// Only progress if some state is registered.
		if s.registered {
			if !s.concluded {
				// Check isFinal and timeout which can indicate a concluded event.
				if d.IsFinal || s.timeoutElapsed() {
					s.concluded = true // ConcludedEvent is only emitted once.
					return s.makeConcludedEvent(d)
				}

				// Otherwise, check state change which indicates a registered event.
				if !d.Equal(s.prevState) {
					s.prevState = *d
					return s.makeRegisteredEvent(d)
				}
			} else {
				// There will be no further events because the ConcludedEvent got already returned.
				s.err <- fmt.Errorf("already concluded")
				return nil
			}
		}

		select {
		case <-s.ctx.Done():
			s.err <- s.ctx.Err()
			return nil
		case <-time.After(s.adjudicator.polling):
		}
	}
}

// Err returns the error status of the subscription. After Next returns nil,
// Err should be checked for an error.
func (s *EventSubscription) Err() error {
	return <-s.err
}

// Close closes the subscription.
func (s *EventSubscription) Close() error {
	s.once.Do(func() { close(s.err) })
	return nil
}

// makeRegisteredEvent returns a new registered event dependent on the given state.
func (s *EventSubscription) makeRegisteredEvent(d *adj.StateReg) channel.AdjudicatorEvent {
	s.timeout = s.convertStateTimeout(d)
	state := d.State.CoreState()
	cID := state.ID
	v := state.Version

	return channel.NewRegisteredEvent(cID, s.timeout, v, state, nil)
}

// makeConcludedEvent returns a new concluded or registered event dependent on the given state and timeout.
func (s *EventSubscription) makeConcludedEvent(d *adj.StateReg) channel.AdjudicatorEvent {
	s.timeout = s.convertStateTimeout(d)
	state := d.State.CoreState()
	cID := state.ID
	v := state.Version

	return channel.NewConcludedEvent(cID, s.timeout, v)
}

// timeoutElapsed evaluates once, if the last recorded timeout is concluded to implicitly indicate a concluded event.
func (s *EventSubscription) timeoutElapsed() bool {
	if s.timeout == nil {
		return false
	}
	return s.timeout.IsElapsed(context.Background())
}

// convertStateTimeout converts the timeout to the client's system time.
func (s *EventSubscription) convertStateTimeout(d *adj.StateReg) *Timeout {
	timeoutDuration := d.Timeout.Time().Sub(d.State.Now.Time())
	return MakeTimeout(timeoutDuration, s.adjudicator.polling)
}

// getState queries the current channel state from the ledger.
// If there is no state available yet pass.
func (s *EventSubscription) getState() (*adj.StateReg, error) {
	ch := s.channelID
	d, err := s.adjudicator.binding.StateReg(ch)

	// Check fist time registration.
	if err != nil {
		if s.registered || !fabclient.IsChannelUnknown(err) {
			return nil, err
		}
	} else if !s.registered {
		s.registered = true
	}

	return d, nil
}
