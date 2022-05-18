package channel

import (
	"context"
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/channel/binding"
	"perun.network/go-perun/channel"
	"time"
)

const (
	defaultAdjPollingInterval = 1 * time.Second
)

// Adjudicator provides methods for dispute resolution on the ledger.
type Adjudicator struct {
	binding *binding.Adjudicator // binding gives access to the Adjudicator contract.
	polling time.Duration        // The polling interval for event subscription.
}

type AdjudicatorOpt func(*Adjudicator)

func AdjudicatorPollingIntervalOpt(d time.Duration) AdjudicatorOpt {
	return func(a *Adjudicator) {
		a.polling = d
	}
}

func NewAdjudicator(adjContract *binding.Adjudicator, opts ...AdjudicatorOpt) *Adjudicator {
	a := &Adjudicator{
		binding: adjContract,
		polling: defaultAdjPollingInterval,
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
	sigCh, err := adj.ConvertToSignedChannel(req) // Repackaging - TODO: Check for a better way here
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
	sigCh, err := adj.ConvertToSignedChannel(req)
	if err != nil {
		return fmt.Errorf("conversion: %w", err)
	}

	err = a.binding.Register(sigCh)
	if err != nil {
		return fmt.Errorf("concluding: %w", err)
	}

	for {
		_, err = a.binding.Withdraw(req.Params.ID())
		waitFor := time.Second * 1
		// If state is not final (and withdraw is blocked) we receive a ChallengeTimeoutError
		if err != nil {
			if cte, ok := err.(adj.ChallengeTimeoutError); ok {
				timeout := cte.Timeout.(adj.StdTimestamp).Time()
				now := cte.Now.(adj.StdTimestamp).Time()
				waitFor = timeout.Sub(now) // Time until challenge duration is passed.
			}
		} else {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitFor):
		}
	}
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
	sub, err := NewEventSubscription(a, ch)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return sub, nil
}
