// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// MemLedger is a simple in-memory ledger, using Go maps.
// time.Time is used as Timestamps.
type MemLedger struct {
	states   map[channel.ID]*StateReg
	holdings map[string]*big.Int
}

func IDKey(id channel.ID) string {
	return hex.EncodeToString(id[:])
}

func FundingKey(id channel.ID, addr wallet.Address) string {
	return fmt.Sprintf("%x:%s", id, addr)
}

func NewMemLedger() *MemLedger {
	return &MemLedger{
		states:   make(map[channel.ID]*StateReg),
		holdings: make(map[string]*big.Int),
	}
}

func (m *MemLedger) GetState(id channel.ID) (*StateReg, error) {
	s, ok := m.states[id]
	if !ok {
		return nil, &NotFoundError{Key: IDKey(id), Type: "StateReg"}
	}
	return s.Clone(), nil
}

func (m *MemLedger) PutState(s *StateReg) error {
	m.states[s.ID] = s.Clone()
	return nil
}

func (m *MemLedger) GetHolding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	key := FundingKey(id, addr)
	h, ok := m.holdings[key]
	if !ok {
		return nil, &NotFoundError{Key: key, Type: "Holding[*big.Int]"}
	}
	return new(big.Int).Set(h), nil
}

func (m *MemLedger) PutHolding(id channel.ID, addr wallet.Address, holding *big.Int) error {
	m.holdings[FundingKey(id, addr)] = new(big.Int).Set(holding)
	return nil
}

// Now returns time.Now() as a Timestamp
func (m *MemLedger) Now() Timestamp {
	return StdNow()
}
