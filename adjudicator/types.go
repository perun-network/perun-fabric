package adjudicator

import (
	"math/big"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	_ "perun.network/go-perun/backend/ethereum"
)

type (
	Params struct {
		ChallengeDuration uint64        `json:"challengeDuration"`
		Parts             []Address     `json:"parts"`
		Nonce             channel.Nonce `json:"nonce"`
	}

	State struct {
		ID       channel.ID    `json:"id"`
		Version  uint64        `json:"version"`
		Balances []channel.Bal `json:"balances"`
		IsFinal  bool          `json:"final"`
	}

	SignedChannel struct {
		Params Params
		State  State
		Sigs   []wallet.Sig
	}

	// Address is the concrete address type used.
	// Unfortunately, this has to be set because fabric-gateway has strong
	// limitations on what types can cross the chaincode API boundary.
	Address = ethwallet.Address

	StateReg struct {
		State
		Timeout *RegTimestamp
	}

	// RegTimestamp is the concrete timestamp type used in the registry.
	// Unfortunately, this has to be set because fabric-gateway has strong
	// limitations on what types can cross the chaincode API boundary.
	RegTimestamp = StdTimestamp
)

func (p Params) ID() channel.ID {
	return channel.CalcID(p.CoreParams())
}

// CoreParams returns the equivalent representation of p as channel.Params.
// The returned Params is set to have no App, LedgerChannel is set to true and
// VirtualChannel is set to false.
//
// It is not a deep copy, e.g., field Parts references the same participants.
func (p Params) CoreParams() *channel.Params {
	return &channel.Params{
		ChallengeDuration: p.ChallengeDuration,
		Parts:             AsWalletAddresses(p.Parts),
		Nonce:             p.Nonce,
		App:               channel.NoApp(),
		LedgerChannel:     true,
	}
}

func AsWalletAddresses(as []Address) []wallet.Address {
	was := make([]wallet.Address, 0, len(as))
	for _, a := range as {
		was = append(was, &a)
	}
	return was
}

// CoreState returns the equivalent representation of s as channel.State.
// The returned State is set to have no App, contains one asset that is default
// initialized and this first assets' balances are set to the Balances of s.
//
// Use the State returned by CoreState to create or verify signatures with the
// go-perun channel backend.
//
// It is not a deep copy, e.g., field Balances references the same balances
// slice.
func (s State) CoreState() *channel.State {
	return &channel.State{
		ID:      s.ID,
		Version: s.Version,
		App:     channel.NoApp(),
		Allocation: channel.Allocation{
			Assets:   []channel.Asset{channel.NewAsset()},
			Balances: channel.Balances{s.Balances},
		},
	}
}

func (s State) Total() channel.Bal {
	total := new(big.Int)
	for _, bal := range s.Balances {
		total.Add(total, bal)
	}
	return total
}

func (s State) Clone() State {
	bals := channel.CloneBals(s.Balances)
	s.Balances = bals
	// Other fields are value types, so done
	return s
}

func (s *StateReg) Clone() *StateReg {
	return &StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone().(*RegTimestamp),
	}
}
