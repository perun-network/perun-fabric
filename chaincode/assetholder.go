// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"perun.network/go-perun/channel"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type AssetHolder struct {
	contractapi.Contract
}

func (AssetHolder) contract(ctx contractapi.TransactionContextInterface) *adj.AssetHolder {
	return adj.NewAssetHolder(NewStubLedger(ctx))
}

func (h *AssetHolder) Deposit(ctx contractapi.TransactionContextInterface,
	id channel.ID, part Address, amountStr string) error {
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}
	return h.contract(ctx).Deposit(id, &part, amount)
}

func (h *AssetHolder) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, part Address) (string, error) {
	return stringWithErr(h.contract(ctx).Holding(id, &part))
}

func (h *AssetHolder) TotalHolding(ctx contractapi.TransactionContextInterface,
	id channel.ID, parts []Address) (string, error) {
	wparts := AsWalletAddresses(parts)
	return stringWithErr(h.contract(ctx).TotalHolding(id, wparts))
}

func (h *AssetHolder) Withdraw(ctx contractapi.TransactionContextInterface,
	id channel.ID, part Address) (string, error) {
	return stringWithErr(h.contract(ctx).Withdraw(id, &part))
}
