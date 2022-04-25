package channel

import (
	"context"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	asset "github.com/perun-network/perun-fabric/tests/assetholder" // TODO: Is "program" not an importable package
	"log"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	"time"
)

const defaultPollingInterval = 1 * time.Second // TODO: Consider a better place for the constant

type Funder struct {
	contract *asset.AssetHolder
	polling  time.Duration
}

// NewFunder returns a new Funder.
func NewFunder(n *client.Network) *Funder { // TODO: Make polling interval adjustable
	return &Funder{
		contract: asset.NewAssetHolder(n), // network is already a gateway connection of a specific acc. Maybe gen gateway here... ?
		polling:  defaultPollingInterval,
	}
}

// Fund deposits funds according to the specified funding request and waits until the funding is complete.
func (f *Funder) Fund(ctx context.Context, req pchannel.FundingReq) error {
	// Get Funding args.
	id := req.Params.ID()
	funding := req.Agreement[0][req.Idx] // TODO: Figure out how to get right balance (of callee)

	// Make deposit.
	err := f.contract.Deposit(id, funding)
	if err != nil {
		return err
	}

	// Wait for Funding completion.
	return f.awaitFundingComplete(ctx, req)
}

// awaitFundingComplete blocks until the funding of the specified channel is complete.
func (f *Funder) awaitFundingComplete(ctx context.Context, req pchannel.FundingReq) error {
	total := req.State.Allocation.Sum()[0] // [0] because we are only supporting one asset.
	for {
		funded := big.NewInt(0) // TODO: We need an Asset here. Temporarily using big int's.
		for i := range req.Params.Parts {
			_funded, err := f.contract.Holding(req.Params.ID(), req.Params.Parts[i]) // TODO: We could use "TotalHolding" here to replace the loop
			if err != nil {
				log.Printf("Warning: Error querying deposit: %v\n", err)
			}
			funded = big.NewInt(0).Add(funded, _funded)
		}

		if funded.Cmp(total) == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(f.polling):
		}
	}

}
