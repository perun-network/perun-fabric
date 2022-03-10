package adjudicator

import (
	"perun.network/go-perun/backend/ethereum/wallet"
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

func AsWalletAddresses(as []Address) []wallet.Address {
	was := make([]wallet.Address, 0, len(as))
	for _, a := range as {
		was = append(was, &a)
	}
	return was
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
