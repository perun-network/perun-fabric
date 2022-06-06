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
	"sync"
	"time"

	"perun.network/go-perun/channel"
)

// EventSubscription provides methods for consuming channel events.
type EventSubscription struct {
	adjudicator *Adjudicator  // adjudicator is the referenced adjudicator instance.
	channelID   channel.ID    // channelID is the channel identifier
	prevState   adj.StateReg  // prevState is the previous channel state
	timeout     *Timeout      // timeout is the current Event timeout
	concluded   bool          // concluded indicates if a concluded event was created
	closed      chan struct{} // closed signals if the sub is closed
	err         chan error    // err forwards errors during event parsing
	once        sync.Once     // once used to close channels
}

func NewEventSubscription(a *Adjudicator, ch channel.ID) (*EventSubscription, error) {
	return &EventSubscription{
		adjudicator: a,
		channelID:   ch,
		closed:      make(chan struct{}),
		err:         make(chan error, 1),
		prevState:   adj.StateReg{},
		timeout:     nil,
	}, nil
}

// Next returns the most recent or next future event. If the subscription is
// closed or any other error occurs, it returns nil.
func (s *EventSubscription) Next() channel.AdjudicatorEvent {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventChan := make(chan channel.AdjudicatorEvent)
	errChan := make(chan error)

	go func() {
		for {
			ch := s.channelID
			d, err := s.adjudicator.binding.StateReg(ch) // Read state
			if err != nil {
				errChan <- err
				return
			}

			if !s.concluded {
				// Check isFinal and timeout which indicates a concluded event.
				//TODO/QUESTION: After the concluded event there will never be another event. How do we handle this?
				if d.IsFinal || s.timeoutElapsed() {
					s.concluded = true
					eventChan <- s.makeConcludedEvent(d)
					return
				}

				// Check state change which indicates a registered event.
				if !d.Equal(s.prevState) {
					s.prevState = *d
					eventChan <- s.makeRegisteredEvent(d)
					return
				}
			} else {
				errChan <- fmt.Errorf("subscription: channel already concluded")
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(s.adjudicator.polling):
			}
		}
	}()

	select {
	case <-s.closed:
		return nil
	case err := <-errChan:
		s.err <- err
		return nil
	case e := <-eventChan:
		return e
	}
}

// Err returns the error status of the subscription. After Next returns nil,
// Err should be checked for an error.
func (s *EventSubscription) Err() error {
	select {
	case <-s.closed:
		return nil // TODO/QUESTION: Discuss nil
	default:
		return <-s.err
	}
}

// Close closes the subscription.
func (s *EventSubscription) Close() error {
	s.once.Do(func() { close(s.closed) })
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
