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
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"sync"
	"time"

	"perun.network/go-perun/channel"
)

// EventSubscription provides methods for consuming channel events.
type EventSubscription struct {
	adjudicator *Adjudicator  // binding is the referenced adjudicator instance.
	channelID   channel.ID    // channelID is the channel identifier
	prev        adj.StateReg  // prev is the previous channel state
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
		prev:        adj.StateReg{},
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

			if !d.Equal(s.prev) {
				s.prev = *d
				eventChan <- s.makeEvent(d)
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
		return nil
	default:
		return <-s.err
	}
}

// Close closes the subscription.
func (s *EventSubscription) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

func (s *EventSubscription) makeEvent(d *adj.StateReg) channel.AdjudicatorEvent {
	state := d.State.CoreState()
	cID := state.ID
	v := state.Version
	t := d.Timeout.(adj.StdTimestamp).Time()
	timeout := makeTimeout(t, s.adjudicator.polling)

	if d.IsFinal {
		return channel.NewConcludedEvent(cID, timeout, v)
	}
	return channel.NewRegisteredEvent(cID, timeout, v, state, nil)
}
