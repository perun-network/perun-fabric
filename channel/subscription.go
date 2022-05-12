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
	evSub <-chan *client.ChaincodeEvent
	once  sync.Once
}

func NewEventSubscription(ctx context.Context, network client.Network, adjudicator string) (*EventSubscription, error) {
	ce, err := network.ChaincodeEvents(ctx, adjudicator)
	if err != nil {
		return nil, fmt.Errorf("concluding: %w", err) // TODO: error description
	}

	return &EventSubscription{
		evSub: ce,
	}, nil
}

// Next returns the most recent or next future event. If the subscription is
// closed or any other error occurs, it returns nil.
func (s *EventSubscription) Next() channel.AdjudicatorEvent {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			e := <-s.evSub // TODO: Query state from here
			event, err := s.makeEvent(e)

			//if err != nil && err.Error() != "Unknown dispute: query wasm contract failed" {
			//	errChan <- err
			//	return
			//}

			//if !d.Equal(s.prev) {
			//	s.prev = d
			//	eventChan <- s.makeEvent(d)
			//	return
			// }

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
	s.once.Do(func() { close(s.evSub) })
	return nil
}

func (s *EventSubscription) makeEvent(e client.ChaincodeEvent) (channel.AdjudicatorEvent, error) {
	// TODO: Parsing event logic
	event := channel.NewRegisteredEvent(cID, timeout, v, state, nil)
	return event, nil
}
