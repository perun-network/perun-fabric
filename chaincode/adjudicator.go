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
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"math/big"
	"perun.network/go-perun/channel"
)

// Adjudicator is the chaincode that implements the adjudicator.
type Adjudicator struct {
	contractapi.Contract
}

func (Adjudicator) contract(ctx contractapi.TransactionContextInterface) *adj.Adjudicator {
	return adj.NewAdjudicator(ctx.GetStub().GetChannelID(), NewStubLedger(ctx), NewStubAsset(ctx))
}

// Deposit unmarshalls the given arguments to forward the deposit request.
func (a *Adjudicator) Deposit(ctx contractapi.TransactionContextInterface,
	chID channel.ID, partStr string, amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10) //nolint:gomnd
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return err
	}

	return a.contract(ctx).Deposit(adj.AccountID(calleeID), chID, part, amount)
}

// Holding unmarshalls the given arguments to forward the holding request.
// It returns the holding amount as a marshalled (string) *big.Int.
func (a *Adjudicator) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string) (string, error) {
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).Holding(id, part))
}

// TotalHolding unmarshalls the given arguments to forward the total holding request.
// It returns the sum of all holding amount of the given participants as a marshalled (string) *big.Int.
func (a *Adjudicator) TotalHolding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partsStr string) (string, error) {
	parts, err := UnmarshalAddresses(partsStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).TotalHolding(id, parts))
}

// Register unmarshalls the given argument to forward the register request.
func (a *Adjudicator) Register(ctx contractapi.TransactionContextInterface,
	chStr string) error {
	var ch adj.SignedChannel
	if err := json.Unmarshal([]byte(chStr), &ch); err != nil {
		return err
	}
	return a.contract(ctx).Register(&ch)
}

// StateReg unmarshalls the given argument to forward the state reg request.
// It returns the retrieved state reg marshalled as string.
func (a *Adjudicator) StateReg(ctx contractapi.TransactionContextInterface,
	id channel.ID) (string, error) {
	reg, err := a.contract(ctx).StateReg(id)
	if err != nil {
		return "", err
	}
	regJSON, err := json.Marshal(reg)
	return string(regJSON), err
}

// Withdraw unmarshalls the given argument to forward the withdrawal request.
// It returns the withdrawal amount as a marshalled (string) *big.Int.
func (a *Adjudicator) Withdraw(ctx contractapi.TransactionContextInterface,
	reqStr string) (string, error) {
	var req adj.SignedWithdrawReq
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).Withdraw(req))
}

// MintToken unmarshalls the given argument to forward the minting request.
// The callee is derived from the transaction context.
func (a *Adjudicator) MintToken(ctx contractapi.TransactionContextInterface,
	amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10) //nolint:gomnd
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Mint(adj.AccountID(calleeID), amount)
	if err != nil {
		return err
	}
	return nil
}

// BurnToken unmarshalls the given argument to forward the burning request.
// The callee is derived from the transaction context.
func (a *Adjudicator) BurnToken(ctx contractapi.TransactionContextInterface,
	amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10) //nolint:gomnd
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Burn(adj.AccountID(calleeID), amount)
	if err != nil {
		return err
	}
	return nil
}

// TransferToken unmarshalls the given arguments to forward the token transfer request.
// The sender of the tokens is derived from the transaction context.
func (a *Adjudicator) TransferToken(ctx contractapi.TransactionContextInterface,
	receiverStr string, amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	receiverID, err := UnmarshalID(receiverStr)
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10) //nolint:gomnd
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Transfer(adj.AccountID(calleeID), receiverID, amount)
	if err != nil {
		return err
	}
	return nil
}

// TokenBalance unmarshalls the given argument to forward the token balance request.
// It returns the balance as a marshalled (string) *big.Int.
func (a *Adjudicator) TokenBalance(ctx contractapi.TransactionContextInterface,
	id string) (string, error) {
	idToCheck, err := UnmarshalID(id)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).BalanceOfID(idToCheck))
}
