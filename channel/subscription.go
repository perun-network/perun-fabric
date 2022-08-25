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
	adjudicator *Adjudicator // adjudicator is the referenced adjudicator instance.
	channelID   channel.ID   // channelID is the channel identifier.
	prevState   adj.StateReg // prevState is the previous channel state.
	timeout     *Timeout     // timeout is the current Event timeout.
	concluded   bool         // concluded indicates if a concluded event was created.
	registered  bool         // registered indicates if any state has been registered on-chain.
	err         chan error   // err forwards errors during event parsing.
	quit        chan bool    // quit indicates that the sub got closed.
	closed      bool         // closed indicates that the channel is closed.
	once        sync.Once    // once used to close() channels.
	mtx         sync.Mutex   // mtx secures against a Close() call during the evaluation of detectEvent().
}

// NewEventSubscription generates a subscriber on the given channel.
func NewEventSubscription(a *Adjudicator, ch channel.ID) (*EventSubscription, error) {
	return &EventSubscription{
		adjudicator: a,
		channelID:   ch,
		err:         make(chan error, 1),
		quit:        make(chan bool, 1),
		prevState:   adj.StateReg{},
		timeout:     nil,
	}, nil
}

// Next returns the most recent or next future event. If the subscription is
// closed or any other error occurs, it returns nil.
func (s *EventSubscription) Next() channel.AdjudicatorEvent {
	for {
		event, err := s.detectEvent()
		if err != nil {
			return nil // Err is available in err chan if subscription is not closed.
		} else if event != nil {
			return event
		}

		select {
		case <-s.quit:
			return nil
		case <-time.After(s.adjudicator.polling):
		}
	}
}

// Err returns the error status of the subscription. After Next returns nil,
// Err should be checked for an error.
func (s *EventSubscription) Err() error {
	if s.closed {
		return nil
	}
	return <-s.err
}

// Close closes the subscription.
func (s *EventSubscription) Close() error {
	s.once.Do(func() {
		s.mtx.Lock()
		defer s.mtx.Unlock()

		s.closed = true
		close(s.quit)
		close(s.err)
	})
	return nil
}

// detectEvent compares the previous and current state of the channel to derive new chain events.
func (s *EventSubscription) detectEvent() (channel.AdjudicatorEvent, error) {
	// Lock to prevent closing during evaluation.
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Abort if subscription got closed.
	if s.closed {
		return nil, fmt.Errorf("subscription closed")
	}

	// Get the on chain state.
	d, err := s.fetchState()
	if err != nil {
		s.err <- err
		return nil, err
	}

	// Only progress if some state is registered.
	if s.registered { //nolint:nestif
		if !s.concluded {
			// If channel isFinal or the timeout elapsed the channel is concluded.
			if d.IsFinal || s.timeoutElapsed() {
				s.concluded = true // ConcludedEvent is only emitted once.
				return s.makeConcludedEvent(d), nil
			}

			// Otherwise, check state change which indicates a registered event.
			if !d.Equal(s.prevState) {
				s.prevState = *d
				return s.makeRegisteredEvent(d), nil
			}
		} else {
			// There will be no further events because the ConcludedEvent got already returned.
			err = fmt.Errorf("already concluded")
			s.err <- err
			return nil, err
		}
	}
	return nil, nil
}

// makeRegisteredEvent returns a new registered event dependent on the given state.
func (s *EventSubscription) makeRegisteredEvent(d *adj.StateReg) channel.AdjudicatorEvent {
	s.timeout = MakeTimeout(d.Timeout.Time(), s.adjudicator.polling)
	state := d.State.CoreState()
	cID := state.ID
	v := state.Version

	return channel.NewRegisteredEvent(cID, s.timeout, v, state, nil)
}

// makeConcludedEvent returns a new concluded or registered event dependent on the given state and timeout.
func (s *EventSubscription) makeConcludedEvent(d *adj.StateReg) channel.AdjudicatorEvent {
	s.timeout = MakeTimeout(d.Timeout.Time(), s.adjudicator.polling)
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

// fetchState queries the current channel state from the ledger.
// If there is no state available yet, pass.
func (s *EventSubscription) fetchState() (*adj.StateReg, error) {
	ch := s.channelID
	d, err := s.adjudicator.binding.StateReg(ch)

	// Check fist time registration.
	if err != nil {
		if s.registered || !fabclient.IsChannelUnknownErr(err) {
			return nil, err
		}
	} else if !s.registered {
		s.registered = true
	}

	return d, nil
}
