package channel

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	chaincode "github.com/perun-network/perun-fabric/chaincode"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"time"
)

// Adjudicator provides methods for dispute resolution on the ledger.
type Adjudicator struct {
	contract *chaincode.Adjudicator
	polling  time.Duration
}

type AdjudicatorOpt func(*Adjudicator)

func AdjudicatorPollingIntervalOpt(d time.Duration) AdjudicatorOpt {
	return func(a *Adjudicator) {
		a.polling = d
	}
}

func NewAdjudicator(c contractapi.Contract, opts ...AdjudicatorOpt) *Adjudicator {
	a := &Adjudicator{
		contract: chaincode.NewAdjudicator(c),
		polling:  defaultPollingInterval,
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
	return a.dispute(ctx, req)
}

// Withdraw concludes and withdraws the registered state, so that the
// final outcome is set on the asset holders and funds are withdrawn.
// If the channel has locked funds in sub-channels, the states of the
// corresponding sub-channels need to be supplied additionally.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	if len(subStates) > 0 {
		return fmt.Errorf("subchannels not supported")
	}
	err := a.conclude(ctx, req)
	if err != nil {
		return fmt.Errorf("concluding: %w", err)
	}
	return a.withdraw(ctx, req)
}

// Progress progresses the state of a previously registered channel on-chain.
// The signatures for the old state can be nil as the state is already
// registered on the adjudicator.
func (a *Adjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	return fmt.Errorf("unsupported")
}

// Subscribe returns an AdjudicatorEvent subscription.
//
// The context should only be used to establish the subscription. The
// framework will call Close on the subscription once the respective channel
// controller shuts down.
func (a *Adjudicator) Subscribe(ctx context.Context, ch channel.ID) (channel.AdjudicatorSubscription, error) {
	sub := NewEventSubscription(a, ch)
	return sub, nil
}
