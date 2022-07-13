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
	return adj.NewAdjudicator(NewStubLedger(ctx), NewStubAsset(ctx))
}

func (a *Adjudicator) Deposit(ctx contractapi.TransactionContextInterface,
	chID channel.ID, partStr string, amountStr string) error {
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return err
	}
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	return a.contract(ctx).Deposit(clientID, chID, part, amount)
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
	part, err := UnmarshalAddress(partStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).Withdraw(id, part))
}

func (a *Adjudicator) MintToken(ctx contractapi.TransactionContextInterface,
	addrStr string, amountStr string) error {
	addr, err := UnmarshalAddress(addrStr)
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	err = a.contract(ctx).Mint(id, addr, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) BurnToken(ctx contractapi.TransactionContextInterface,
	addrStr string, amountStr string) error {
	addr, err := UnmarshalAddress(addrStr)
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	err = a.contract(ctx).Burn(id, addr, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) TokenToAddressTransfer(ctx contractapi.TransactionContextInterface,
	senderAddrStr string, receiverAddrStr string, amountStr string) error {
	senderAddr, err := UnmarshalAddress(senderAddrStr)
	if err != nil {
		return err
	}

	receiverAddr, err := UnmarshalAddress(receiverAddrStr)
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	err = a.contract(ctx).AddressToAddressTransfer(id, senderAddr, receiverAddr, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) TokenToChannelTransfer(ctx contractapi.TransactionContextInterface,
	senderAddrStr string, receiver channel.ID, amountStr string) error {
	senderAddr, err := UnmarshalAddress(senderAddrStr)
	if err != nil {
		return err
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return fmt.Errorf("parsing big.Int string %q failed", amountStr)
	}

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	err = a.contract(ctx).AddressToChannelTransfer(id, senderAddr, receiver, amount)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) TokenBalance(ctx contractapi.TransactionContextInterface,
	addrStr string) (string, error) {
	addr, err := UnmarshalAddress(addrStr)
	if err != nil {
		return "", err
	}
	return stringWithErr(a.contract(ctx).BalanceOfAddress(addr))
}

func (a *Adjudicator) RegisterAddress(ctx contractapi.TransactionContextInterface,
	addr string) error {
	part, err := UnmarshalAddress(addr)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	err = a.contract(ctx).RegisterAddress(clientID, part)
	if err != nil {
		return err
	}
	return nil
}

func (a *Adjudicator) GetAddressIdentity(ctx contractapi.TransactionContextInterface,
	addrStr string) (string, error) {
	addr, err := UnmarshalAddress(addrStr)
	if err != nil {
		return "", err
	}

	id, err := a.contract(ctx).GetAddressIdentity(addr)
	if err != nil {
		return id, err
	}
	return id, nil
}
