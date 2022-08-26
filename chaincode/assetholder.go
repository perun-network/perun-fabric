// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chaincode

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"perun.network/go-perun/channel"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

// AssetHolder is the chaincode that implements the AssetHolder.
// This skips the checks of the Adjudicator and uses a different ledger to store data.
// Hence, it is intended for directly testing the AssetHolder chaincode.
type AssetHolder struct {
	contractapi.Contract
}

func (AssetHolder) contract(ctx contractapi.TransactionContextInterface) *adj.AssetHolder {
	return adj.NewAssetHolder(NewStubLedger(ctx))
}

// Deposit unmarshalls the given arguments to forward the deposit request.
func (h *AssetHolder) Deposit(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string, amountStr string) error {
	amount, ok := new(big.Int).SetString(amountStr, 10) //nolint:gomnd
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return err
	}
	return h.contract(ctx).Deposit(id, part, amount)
}

// Holding unmarshalls the given arguments to forward the holding request.
// It returns the holding amount as a marshalled (string) *big.Int.
func (h *AssetHolder) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string) (string, error) {
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(h.contract(ctx).Holding(id, part))
}

// TotalHolding unmarshalls the given arguments to forward the total holding request.
// It returns the sum of all holding amount of the given participants as a marshalled (string) *big.Int.
func (h *AssetHolder) TotalHolding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partsStr string) (string, error) {
	parts, err := UnmarshalAddresses(partsStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(h.contract(ctx).TotalHolding(id, parts))
}

// Withdraw unmarshalls the given argument to forward the withdrawal request.
// It returns the withdrawal amount as a marshalled (string) *big.Int.
func (h *AssetHolder) Withdraw(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string) (string, error) {
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(h.contract(ctx).Withdraw(id, part))
}
