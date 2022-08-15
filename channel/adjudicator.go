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
	"github.com/hyperledger/fabric-gateway/pkg/client"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/channel/binding"
	fabclient "github.com/perun-network/perun-fabric/client"
	"perun.network/go-perun/channel"
	"time"
)

const (
	defaultAdjPollingInterval = 1 * time.Second
)

// Adjudicator provides methods for dispute resolution on the ledger.
type Adjudicator struct {
	binding  *binding.Adjudicator // binding gives access to the Adjudicator contract.
	polling  time.Duration        // The polling interval for event subscription.
	receiver string               // The fabric id of the receiver of the funds for withdrawal.
}

// AdjudicatorOpt allows to extend the Adjudicator constructor.
type AdjudicatorOpt func(*Adjudicator)

// WithSubPollingInterval overwrites the polling interval for the Adjudicators' event subscription.
func WithSubPollingInterval(d time.Duration) AdjudicatorOpt {
	return func(a *Adjudicator) {
		a.polling = d
	}
}

// NewAdjudicator generates an Adjudicator and requires to preset the fabric ID used for withdrawal.
func NewAdjudicator(network *client.Network, chaincode string, withdrawTo string, opts ...AdjudicatorOpt) *Adjudicator {
	a := &Adjudicator{
		binding:  binding.NewAdjudicatorBinding(network, chaincode),
		polling:  defaultAdjPollingInterval,
		receiver: withdrawTo,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Register registers the given ledger channel state on-chain.
// If the channel has locked funds into sub-channels, the corresponding
// signed sub-channel states must be provided.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	if len(subChannels) > 0 {
		return fmt.Errorf("subchannels not supported")
	}
	sigCh, err := adj.ConvertToSignedChannel(req)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	return a.binding.Register(sigCh)
}

// Withdraw concludes and withdraws the registered state, so that the
// final outcome is set on the asset holders and funds are withdrawn.
// If the channel has locked funds in sub-channels, the states of the
// corresponding sub-channels need to be supplied additionally.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	if len(subStates) > 0 {
		return fmt.Errorf("subchannels not supported")
	}
	channelID := req.Tx.ID

	// For withdrawing there must be at least one registered state.
	if req.Tx.IsFinal { //nolint:nestif
		// Ensure there is a registered state.
		err := a.ensureRegistered(ctx, req, nil)
		if err != nil {
			return err
		}
	} else {
		// Dispute case: There must be a registered state.
		reg, err := a.binding.StateReg(channelID)
		if err != nil {
			return err
		}

		if reg.Version != req.Tx.Version {
			return fmt.Errorf("invalid adjudicator request")
		}

		timeout := MakeTimeout(reg.Timeout.Time(), a.polling)
		err = timeout.Wait(ctx)
		if err != nil {
			return err
		}
	}

	// Concluded (or waited for challenge duration in case of dispute)
	// Withdraw funds.
	withdrawReq, err := adj.SignWithdrawRequest(req.Acc, channelID, a.receiver)
	if err != nil {
		return err
	}
	_, err = a.binding.Withdraw(*withdrawReq)
	if err != nil {
		return err
	}
	return nil
}

// Progress progresses the state of a previously registered channel on-chain.
// The signatures for the old state can be nil as the state is already
// registered on the adjudicator.
func (a *Adjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	return fmt.Errorf("unsupported")
}

// Subscribe returns an AdjudicatorEvent subscription.
// The context should only be used to establish the subscription. The
// framework will call Close on the subscription once the respective channel
// controller shuts down.
func (a *Adjudicator) Subscribe(ctx context.Context, ch channel.ID) (channel.AdjudicatorSubscription, error) {
	sub, err := NewEventSubscription(a, ch)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return sub, nil
}

// ensureRegistered tries to register the state given in AdjudicatorReq.
// If the channel is already registered or the registration is successful ensureRegisters returns nil.
// Otherwise, an error is returned.
func (a *Adjudicator) ensureRegistered(ctx context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	err := a.Register(ctx, req, subChannels)

	// If an underfunded error is returned:
	// Either the clients AdjudicatorReq contains a faulty balance which we do not expect at this point.
	// Or the other party successfully called withdraw which lowered the total holdings in the channel.
	// Hence, the other party must have had registered before.
	if fabclient.IsUnderfundedErr(err) {
		return nil
	}
	return err
}
