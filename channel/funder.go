package channel

import (
	"context"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	assset "github.com/perun-network/perun-fabric/chaincode"
	"log"
	"math/big"
	pchannel "perun.network/go-perun/channel"
	"time"
)

const defaultPollingInterval = 1 * time.Second // TODO: Consider a better place for the constant

type Funder struct {
	contract *assset.AssetHolder
	tctx     *contractapi.TransactionContext
	polling  time.Duration
}

// NewFunder returns a new Funder.
func NewFunder(c contractapi.Contract, tctx *contractapi.TransactionContext) *Funder { // TODO: Make polling interval adjustable
	return &Funder{
		contract: assset.NewAssetHolder(c), // Network is already a gateway connection of a specific acc. Maybe gen gateway here... ?
		tctx:     tctx,                     // Created when a smart contract is deployed to a channel and made available to every subsequent transaction invocation.
		polling:  defaultPollingInterval,
	}
}

// Fund deposits funds according to the specified funding request and waits until the funding is complete.
func (f *Funder) Fund(ctx context.Context, req pchannel.FundingReq) error {
	// Get Funding args.
	id := req.Params.ID()

	if len(req.Agreement) != 1 {
		panic("Funder: Funding request includes less ore more than one asset.")
	}
	assetIndex := 0
	funding := req.Agreement[assetIndex][req.Idx]

	// Make deposit.
	err := f.contract.Deposit(f.tctx, id, funding.String()) // TODO: Check why Deposit required the funding in string format
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
		funded := big.NewInt(0) // TODO: We need to use a asset type here.
		for i := range req.Params.Parts {
			fundedStr, err := f.contract.Holding(f.tctx, req.Params.ID(), req.Params.Parts[i].String()) // TODO: We could use "TotalHolding" here to replace the loop
			if err != nil {
				log.Printf("Warning: Error querying deposit: %v\n", err)
			}

			_funded, _ := big.NewInt(0).SetString(fundedStr, 10) // TODO: Check if we can mitigate this conversion
			funded = funded.Add(funded, _funded)
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
