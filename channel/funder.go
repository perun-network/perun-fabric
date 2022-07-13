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
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/perun-network/perun-fabric/channel/binding"
	"perun.network/go-perun/channel"
	"sync"
	"time"
)

const (
	defaultFunderPollingInterval = 1 * time.Second
)

type Funder struct {
	binding *binding.Adjudicator // binding gives access to the AssetHolder contract.
	polling time.Duration        // The polling interval to wait for complete funding.
	m       sync.Mutex           // m prevents sending parallel transactions.
}

type FunderOpt func(*Funder)

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

	// Wait for Funding completion.
	return f.awaitFundingComplete(ctx, req)
}

// awaitFundingComplete blocks until the funding of the specified channel is complete.
func (f *Funder) awaitFundingComplete(ctx context.Context, req channel.FundingReq) error {
	total := req.State.Allocation.Sum()[0] // [0] because we are only supporting one asset.
	for {
		funded, err := f.binding.TotalHolding(req.State.ID, req.Params.Parts)
		if err != nil {
			return err
		}

		if funded.Cmp(total) >= 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(f.polling):
		}
	}
}
