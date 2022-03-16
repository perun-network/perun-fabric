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

type StubLedger struct {
	Stub shim.ChaincodeStubInterface
}

func NewStubLedger(ctx contractapi.TransactionContextInterface) *StubLedger {
	return &StubLedger{Stub: ctx.GetStub()}
}

func (l *StubLedger) GetState(id channel.ID) (*adj.StateReg, error) {
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

func (l *StubLedger) GetHolding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	key := HoldingKey(id, addr)
	srb, err := l.Stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return nil, &adj.NotFoundError{Key: key, Type: "Holding[*big.Int]"}
	}

	return new(big.Int).SetBytes(srb), nil
}

func (l *StubLedger) PutHolding(id channel.ID, addr wallet.Address, holding *big.Int) error {
	key := HoldingKey(id, addr)
	if err := l.Stub.PutState(key, holding.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// maxNowDiff is the maximum allowed difference of a transaction's timestamp to
// be considered the current block time.
const maxNowDiff = 3 * time.Second

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
	return adj.StdTimestamp(now)
}

func absDuration(d time.Duration) time.Duration {
	if d >= 0 {
		return d
	}
	return -d
}

const orgPrefix = "network.perun."

func StateRegKey(id channel.ID) string {
	return orgPrefix + "StateReg:" + adj.IDKey(id)
}

func HoldingKey(id channel.ID, addr wallet.Address) string {
	return orgPrefix + "Holding:" + adj.FundingKey(id, addr)
}
