// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"math/big"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type AssetHolder struct {
	contractapi.Contract
}

func (h *AssetHolder) contract(ctx contractapi.TransactionContextInterface) *adj.AssetHolder {
	return adj.NewAssetHolder(NewStubLedger(ctx))
}

func (h *AssetHolder) Deposit(ctx contractapi.TransactionContextInterface,
	id channel.ID, part wallet.Address, amount *big.Int) error {
	return h.contract(ctx).Deposit(id, part, amount)
}

func (h *AssetHolder) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, part wallet.Address) (*big.Int, error) {
	return h.contract(ctx).Holding(id, part)
}

func (h *AssetHolder) TotalHolding(ctx contractapi.TransactionContextInterface,
	params *channel.Params) (*big.Int, error) {
	return h.contract(ctx).TotalHolding(params)
}

func (h *AssetHolder) Withdraw(ctx contractapi.TransactionContextInterface,
	id channel.ID, part wallet.Address) (*big.Int, error) {
	return h.contract(ctx).Withdraw(id, part)
}
