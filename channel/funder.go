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
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/perun-network/perun-fabric/channel/binding"
	"perun.network/go-perun/channel"
	"sync"
	"time"
)

const (
	defaultFunderPollingInterval = 1 * time.Second
)

// Funder provides functionality for channel funding.
type Funder struct {
	binding *binding.Adjudicator // binding gives access to the chaincode.
	polling time.Duration        // The polling interval to wait for complete funding.
	m       sync.Mutex           // m prevents sending parallel transactions.
}

// FunderOpt extends the constructor of Funder.
type FunderOpt func(*Funder)

// WithPollingInterval overwrites the polling interval to check for funding completion.
func WithPollingInterval(d time.Duration) FunderOpt {
	return func(f *Funder) {
		f.polling = d
	}
}

// NewFunder returns a new Funder.
func NewFunder(network *client.Network, chaincode string, opts ...FunderOpt) *Funder {
	f := &Funder{
		binding: binding.NewAdjudicatorBinding(network, chaincode),
		polling: defaultFunderPollingInterval,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Fund deposits funds according to the specified funding request and waits until the funding is complete.
func (f *Funder) Fund(ctx context.Context, req channel.FundingReq) error {
	// Get Funding args.
	id := req.State.ID

	if len(req.Agreement) != 1 {
		panic("Funder: Funding request does not hold one asset.")
	}
	assetIndex := 0
	funding := req.Agreement[assetIndex][req.Idx]
	part := req.Params.Parts[req.Idx]

	// Make deposit.
	f.m.Lock()
	defer f.m.Unlock()
	err := f.binding.Deposit(id, part, funding)
	if err != nil {
		return err
	}

	// Calculate funding timeout.
	challengeDuration := time.Duration(req.Params.ChallengeDuration) * time.Second
	wall := time.Now().UTC().Add(challengeDuration)
	timeout := MakeTimeout(wall, f.polling)

	// Wait for Funding completion.
	return f.awaitFundingComplete(ctx, timeout, req)
}

// awaitFundingComplete blocks until the funding of the specified channel is complete.
func (f *Funder) awaitFundingComplete(ctx context.Context, t *Timeout, req channel.FundingReq) error {
	assetIdx := 0 // We only support one asset.
	total := req.State.Allocation.Sum()[assetIdx]
	for {
		funded, err := f.binding.TotalHolding(req.State.ID, req.Params.Parts)
		if err != nil {
			return err
		}

		// Check if funding completed.
		if funded.Cmp(total) >= 0 {
			return nil
		}

		// Check if funding failed.
		if t.IsElapsed(ctx) {
			otherParty := 1 - req.Idx // We only support the two-party case.
			return channel.NewFundingTimeoutError(
				[]*channel.AssetFundingError{{
					Asset:         channel.Index(assetIdx),
					TimedOutPeers: []channel.Index{otherParty},
				}},
			)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(f.polling):
		}
	}
}
