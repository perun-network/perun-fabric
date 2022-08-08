// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"math/big"
	"perun.network/go-perun/channel"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type Adjudicator struct {
	contractapi.Contract
}

func (Adjudicator) contract(ctx contractapi.TransactionContextInterface) *adj.Adjudicator {
	return adj.NewAdjudicator(ctx.GetStub().GetChannelID(), NewStubLedger(ctx), NewStubAsset(ctx))
}

func (a *Adjudicator) Deposit(ctx contractapi.TransactionContextInterface,
	chID channel.ID, partStr string, amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return err
	}

	return a.contract(ctx).Deposit(calleeID, chID, part, amount)
}

func (a *Adjudicator) Holding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string) (string, error) {
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).Holding(id, part))
}

func (a *Adjudicator) TotalHolding(ctx contractapi.TransactionContextInterface,
	id channel.ID, partsStr string) (string, error) {
	parts, err := UnmarshalAddresses(partsStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).TotalHolding(id, parts))
}

func (a *Adjudicator) Register(ctx contractapi.TransactionContextInterface,
	chStr string) error {
	var ch adj.SignedChannel
	if err := json.Unmarshal([]byte(chStr), &ch); err != nil {
		return err
	}
	return a.contract(ctx).Register(&ch)
}

func (a *Adjudicator) StateReg(ctx contractapi.TransactionContextInterface,
	id channel.ID) (string, error) {
	reg, err := a.contract(ctx).StateReg(id)
	if err != nil {
		return "", err
	}
	regJson, err := json.Marshal(reg)
	return string(regJson), err
}

func (a *Adjudicator) Withdraw(ctx contractapi.TransactionContextInterface,
	id channel.ID, partStr string) (string, error) {
	calleeID, err := ctx.GetClientIdentity().GetID() // TODO: Remove for new Withdraw logic with signed request.
	if err != nil {
		return "", err
	}

	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).Withdraw(calleeID, id, part))
}

func (a *Adjudicator) MintToken(ctx contractapi.TransactionContextInterface,
	amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Mint(calleeID, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) BurnToken(ctx contractapi.TransactionContextInterface,
	amountStr string) error {
	calleeID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Burn(calleeID, amount)
	if err != nil {
		return err
	}
	return nil
}

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

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	err = a.contract(ctx).Transfer(calleeID, receiverID, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) TokenBalance(ctx contractapi.TransactionContextInterface,
	id string) (string, error) {
	idToCheck, err := UnmarshalID(id)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).BalanceOfID(idToCheck))
}
