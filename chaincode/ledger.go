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
	"log"
	"math/big"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

// StubLedger is an on-chain ledger.
type StubLedger struct {
	Stub shim.ChaincodeStubInterface
}

// NewStubLedger returns a ledger that uses the stub of the transaction context for storing information.
func NewStubLedger(ctx contractapi.TransactionContextInterface) *StubLedger {
	return &StubLedger{Stub: ctx.GetStub()}
}

// GetState retrieves the current channel state.
func (l *StubLedger) GetState(id channel.ID) (*adj.StateReg, error) { //nolint:forbidigo
	key := StateRegKey(id)
	srb, err := l.Stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return nil, &adj.NotFoundError{Key: key, Type: "StateReg"}
	}

	var sr adj.StateReg
	return &sr, json.Unmarshal(srb, &sr)
}

// PutState overwrites the current channel state with the given one.
func (l *StubLedger) PutState(sr *adj.StateReg) error {
	srb, err := json.Marshal(sr)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	key := StateRegKey(sr.ID)
	if err := l.Stub.PutState(key, srb); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// GetHolding retrieves the current channel holding of the given address.
func (l *StubLedger) GetHolding(id channel.ID, addr wallet.Address) (*big.Int, error) { //nolint:forbidigo
	key := ChannelHoldingKey(id, addr)
	srb, err := l.Stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return nil, &adj.NotFoundError{Key: key, Type: "Holding[*big.Int]"}
	}

	return new(big.Int).SetBytes(srb), nil
}

// PutHolding overwrites the current address channel holdings with the given holding.
func (l *StubLedger) PutHolding(id channel.ID, addr wallet.Address, holding *big.Int) error {
	key := ChannelHoldingKey(id, addr)
	if err := l.Stub.PutState(key, holding.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// maxNowDiff is the maximum allowed difference of a transaction's timestamp to
// be considered the current block time.
const maxNowDiff = 3 * time.Second

// Now retrieves the transaction timestamp.
func (l *StubLedger) Now() adj.Timestamp {
	pbts, err := l.Stub.GetTxTimestamp()
	if err != nil {
		log.Panicf("error getting transaction timestamp: %v", err)
	}
	now := pbts.AsTime()
	localnow := time.Now()
	if absDuration(now.Sub(localnow)) > maxNowDiff {
		log.Panicf("transaction timestamp (%v) too far off local now (%v)", now, localnow)
	}
	return adj.Timestamp(now)
}

func absDuration(d time.Duration) time.Duration {
	if d >= 0 {
		return d
	}
	return -d
}

const orgPrefix = "network.perun."

// StateRegKey generates the key for storing the channel state on the stub.
func StateRegKey(id channel.ID) string {
	return orgPrefix + "ChannelStateReg:" + adj.IDKey(id)
}

// ChannelHoldingKey generates the key for storing holdings on the stub.
func ChannelHoldingKey(id channel.ID, addr wallet.Address) string {
	return orgPrefix + "ChannelHolding:" + adj.FundingKey(id, addr)
}
