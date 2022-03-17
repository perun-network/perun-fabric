// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"math/big"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	pkgjson "github.com/perun-network/perun-fabric/pkg/json"
	"github.com/perun-network/perun-fabric/tests"
)

const (
	txDeposit      = "Deposit"
	txHolding      = "Holding"
	txTotalHolding = "TotalHolding"
	txRegister     = "Register"
	txStateReg     = "StateReg"
	txWithdraw     = "Withdraw"
)

type Adjudicator struct {
	Contract *client.Contract
}

func NewAdjudicator(network *client.Network) *Adjudicator {
	return &Adjudicator{Contract: network.GetContract(*chainCode)}
}

func (a *Adjudicator) Deposit(id channel.ID, addr wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, addr, amount)
	if err != nil {
		return err
	}
	_, err = a.Contract.SubmitTransaction(txDeposit, args...)
	return err
}

func (a *Adjudicator) Holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.Contract.SubmitTransaction(txHolding, args...))
}

func (a *Adjudicator) TotalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.Contract.SubmitTransaction(txTotalHolding, args...))
}

func (a *Adjudicator) Register(ch *adj.SignedChannel) error {
	arg, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	_, err = a.Contract.SubmitTransaction(txRegister, string(arg))
	return err
}

func (a *Adjudicator) StateReg(id channel.ID) (*adj.StateReg, error) {
	arg, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	regJson, err := a.Contract.SubmitTransaction(txStateReg, string(arg))
	if err != nil {
		return nil, err
	}
	var reg adj.StateReg
	return &reg, json.Unmarshal(regJson, &reg)
}

func (a *Adjudicator) Withdraw(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(a.Contract.SubmitTransaction(txWithdraw, args...))
}
