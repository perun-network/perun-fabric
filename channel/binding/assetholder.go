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

package binding

import (
	"math/big"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	pkgjson "github.com/perun-network/perun-fabric/pkg/json"
)

// AssetHolder wraps a fabric client.Contract to connect to the AssetHolder chaincode.
type AssetHolder struct {
	Contract *client.Contract
}

// NewAssetHolderBinding creates the bindings for the on-chain AssetHolder.
// These are only needed for isolated AssetHolder chaincode testing.
// There is no connection to the Adjudicator here.
func NewAssetHolderBinding(network *client.Network, chainCode string) *AssetHolder {
	return &AssetHolder{Contract: network.GetContract(chainCode)}
}

// Deposit marshals the given parameters and sends a deposits request to the AssetHolder chaincode.
func (ah *AssetHolder) Deposit(id channel.ID, part wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, part, amount)
	if err != nil {
		return err
	}
	_, err = ah.Contract.SubmitTransaction(txDeposit, args...)
	return err
}

// Holding marshals the given parameters and sends a holding request to the AssetHolder chaincode.
// The response contains the current holding of the given address in the channel.
func (ah *AssetHolder) Holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(ah.Contract.SubmitTransaction(txHolding, args...))
}

// TotalHolding marshals the given parameters and sends a total holding request to the AssetHolder chaincode.
// The response contains the sum of the current holdings of the given addresses in the channel.
func (ah *AssetHolder) TotalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(ah.Contract.SubmitTransaction(txTotalHolding, args...))
}

// Withdraw marshals the given parameters and sends a withdrawal request to the AssetHolder chaincode.
// The response contains the amount of funds withdrawn form the channel.
func (ah *AssetHolder) Withdraw(id channel.ID, part wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, part)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(ah.Contract.SubmitTransaction(txWithdraw, args...))
}
