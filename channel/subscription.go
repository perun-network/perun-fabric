//  Copyright 2021 PolyCrypt GmbH
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
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"sync"
	"time"

	"perun.network/go-perun/channel"
)

// EventSubscription provides methods for consuming channel events.
type EventSubscription struct {
	listener <-chan *client.ChaincodeEvent // Listen for events here
	closed   chan struct{}                 // closed signals if the sub is closed
	err      chan error                    // err forwards errors during event parsing
	once     sync.Once                     // once used to close channels
}

func NewEventSubscription(ctx context.Context, ch channel.ID, network client.Network, adjudicator string) (*EventSubscription, error) {
	ce, err := network.ChaincodeEvents(ctx, adjudicator)
	if err != nil {
		return nil, fmt.Errorf("concluding: %w", err) // TODO: error description
	}

	return &EventSubscription{
		listener: ce,
		closed:   make(chan struct{}),
		err:      make(chan error, 1),
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
			e := <-s.listener
			event, err := s.makeEvent(*e)

			if err != nil {
				errChan <- err // Propagate error
				return
			}
			eventChan <- event

			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultPollingInterval):
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

func (s *EventSubscription) makeEvent(e client.ChaincodeEvent) (channel.AdjudicatorEvent, error) {
	// TODO: Parsing event logic
	// event := channel.NewRegisteredEvent(cID, timeout, v, state, nil)
	return nil, nil
}
