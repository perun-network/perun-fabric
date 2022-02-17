// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type (
	// MemLedger is a simple in-memory ledger, using Go maps.
	// time.Time is used as Timestamps.
	MemLedger struct {
		states   map[channel.ID]*StateReg
		holdings map[string]*big.Int
	}

	StdTimestamp time.Time
)

func FundingKey(id channel.ID, addr wallet.Address) string {
	return fmt.Sprintf("%x:%s", id, addr)
}

func StdNow() StdTimestamp { return StdTimestamp(time.Now()) }

func (t StdTimestamp) Time() time.Time { return (time.Time)(t) }

func asTime(t Timestamp) time.Time { return t.(StdTimestamp).Time() }

func (t StdTimestamp) Equal(other Timestamp) bool {
	return t.Time().Equal(asTime(other))
}

func (t StdTimestamp) After(other Timestamp) bool {
	return t.Time().After(asTime(other))
}

func (t StdTimestamp) Before(other Timestamp) bool {
	return t.Time().Before(asTime(other))
}

func NewMemLedger() *MemLedger {
	return &MemLedger{
		states:   make(map[channel.ID]*StateReg),
		holdings: make(map[string]*big.Int),
	}
}

// Clone returns t itself, which is a proper clone. Note that time.Time is to be
// used like a scalar.
func (t StdTimestamp) Clone() Timestamp {
	return t
}

func (m *MemLedger) GetState(id channel.ID) (*StateReg, error) {
	s, ok := m.states[id]
	if !ok {
		return nil, &NotFoundError{Key: hex.EncodeToString(id[:]), Type: "StateReg"}
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

// Now returns time.Now()
func (m *MemLedger) Now() Timestamp {
	return StdNow()
}
