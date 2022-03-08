// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"fmt"
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
	id channel.ID, part adj.Address, amountStr string) error {
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}
	return h.contract(ctx).Deposit(id, &part, amount)
}

func (h *AssetHolder) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, part adj.Address) (string, error) {
	return stringWithErr(h.contract(ctx).Holding(id, &part))
}

func (h *AssetHolder) TotalHolding(ctx contractapi.TransactionContextInterface,
	id channel.ID, parts []adj.Address) (string, error) {
	wparts := make([]wallet.Address, 0, len(parts))
	for _, p := range parts {
		wparts = append(wparts, &p)
	}
	return stringWithErr(h.contract(ctx).TotalHolding(id, wparts))
}

func (h *AssetHolder) Withdraw(ctx contractapi.TransactionContextInterface,
	id channel.ID, part adj.Address) (string, error) {
	return stringWithErr(h.contract(ctx).Withdraw(id, &part))
}
