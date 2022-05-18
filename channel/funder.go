package channel

import (
	"context"
	"github.com/perun-network/perun-fabric/channel/binding"
	"perun.network/go-perun/channel"
	"time"
)

const (
	defaultFunderPollingInterval = 1 * time.Second
)

type Funder struct {
	binding *binding.AssetHolder // binding gives access to the AssetHolder contract.
	polling time.Duration        // The polling interval to wait for complete funding.
}

type FunderOpt func(*Funder)

func FunderPollingIntervalOpt(d time.Duration) FunderOpt {
	return func(f *Funder) {
		f.polling = d
	}
}

// NewFunder returns a new Funder.
func NewFunder(assetHolder *binding.AssetHolder, opts ...FunderOpt) *Funder {
	f := &Funder{
		binding: assetHolder,
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
	id := req.Params.ID()

	if len(req.Agreement) != 1 {
		panic("Funder: Funding request does not hold one asset.")
	}
	assetIndex := 0
	funding := req.Agreement[assetIndex][req.Idx]

	// Make deposit.
	err := f.binding.Deposit(id, funding)
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
		funded, err := f.binding.TotalHolding(req.Params.ID(), req.Params.Parts)
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
