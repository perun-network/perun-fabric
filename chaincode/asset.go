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
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"math/big"
)

// centralBankerID is the user allowed to mint & burn tokens.
const centralBankerID = "eDUwOTo6Q049dXNlcjEsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw=="

// StubAsset is an on-chain asset.
type StubAsset struct {
	Stub shim.ChaincodeStubInterface
}

// NewStubAsset returns an Asset that uses the stub of the transaction context for storing asset holdings.
func NewStubAsset(ctx contractapi.TransactionContextInterface) *StubAsset {
	return &StubAsset{Stub: ctx.GetStub()}
}

// Mint creates the desired amount of token for the given id.
// The id must be the callee of the transaction invoking Mint.
func (s StubAsset) Mint(id adj.AccountID, amount *big.Int) error {
	// Check if id == centralBanker.
	if id != centralBankerID {
		return fmt.Errorf("minter must be central banker")
	}

	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOf(id)
	if err != nil {
		return err
	}
	current.Add(current, amount)
	if err := s.Stub.PutState(string(id), current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// Burn removes the desired amount of token from the given id.
// The id must be the callee of the transaction invoking Burn.
func (s StubAsset) Burn(id adj.AccountID, amount *big.Int) error {
	// Check if id == centralBanker.
	if id != centralBankerID {
		return fmt.Errorf("burner must be central banker")
	}

	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot burn zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOf(id)
	if err != nil {
		return err
	}
	current.Sub(current, amount)

	if current.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("not enought funds to burn the requested amount")
	}

	if err := s.Stub.PutState(string(id), current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// Transfer checks if the proposed transfer is valid and
// transfers the given amount of coins from the sender to the receiver.
// The sender must be the callee of the transaction invoking Transfer.
func (s StubAsset) Transfer(sender adj.AccountID, receiver adj.AccountID, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("cannot transfer negative amount")
	}

	// Check balance of sender.
	senderBal, err := s.BalanceOf(sender)
	if err != nil {
		return err
	}
	if !(senderBal.Cmp(amount) >= 0) {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, err := s.BalanceOf(receiver)
	if err != nil {
		return err
	}

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	if err := s.Stub.PutState(string(sender), senderBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	if err := s.Stub.PutState(string(receiver), receiverBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// BalanceOf returns the amount of tokens the given id holds.
// If the id is unknown, zero is returned.
func (s StubAsset) BalanceOf(id adj.AccountID) (*big.Int, error) {
	srb, err := s.Stub.GetState(string(id))
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).SetBytes(srb), nil
}
