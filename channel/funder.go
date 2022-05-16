package channel

import (
	"context"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/perun-network/perun-fabric/pkg/json"
	"github.com/perun-network/perun-fabric/tests"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"time"
)

// TODO: Consider a better place for these constants.
const (
	txDeposit              = "Deposit"
	txHolding              = "Holding"
	txTotalHolding         = "TotalHolding"
	txWithdraw             = "Withdraw"
	defaultPollingInterval = 1 * time.Second
)

type Funder struct {
	ah      *client.Contract // The AssetHolder contract.
	polling time.Duration    // The polling interval.
}

type FunderOpt func(*Funder)

func FunderPollingIntervalOpt(d time.Duration) FunderOpt {
	return func(f *Funder) {
		f.polling = d
	}
}

// NewFunder returns a new Funder.
func NewFunder(assetHolder *client.Contract, opts ...FunderOpt) *Funder {
	f := &Funder{
		ah:      assetHolder,
		polling: defaultPollingInterval,
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
	funding := req.Agreement[assetIndex][req.Idx] // TODO: Check req.Idx

	// Make deposit.
	err := f.deposit(id, funding)
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
		funded, err := f.totalHolding(req.Params.ID(), req.Params.Parts)
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

func (f *Funder) deposit(id channel.ID, amount *big.Int) error {
	args, err := json.MultiMarshal(id, amount)
	if err != nil {
		return err
	}
	_, err = f.ah.SubmitTransaction(txDeposit, args...)
	return err
}

func (f *Funder) holding(id channel.ID, addr wallet.Address) (*big.Int, error) { // TODO: Remove?
	args, err := json.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(f.ah.SubmitTransaction(txHolding, args...))
}

func (f *Funder) totalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := json.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(f.ah.SubmitTransaction(txTotalHolding, args...))
}

func (f *Funder) withdraw(id channel.ID) (*big.Int, error) { // TODO: Remove?
	args, err := json.MultiMarshal(id)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(f.ah.SubmitTransaction(txWithdraw, args...))
}
