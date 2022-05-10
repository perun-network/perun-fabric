package adjudicator

import (
	"encoding/json"
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type (
	Params struct {
		ChallengeDuration uint64           `json:"challengeDuration"`
		Parts             []wallet.Address `json:"parts"`
		Nonce             channel.Nonce    `json:"nonce"`
	}

	State struct {
		ID       channel.ID    `json:"id"`
		Version  uint64        `json:"version"`
		Balances []channel.Bal `json:"balances"`
		IsFinal  bool          `json:"final"`
	}

	SignedChannel struct {
		Params Params       `json:"params"`
		State  State        `json:"state"`
		Sigs   []wallet.Sig `json:"sigs"`
	}

	StateReg struct {
		State   `json:"state"`
		Timeout Timestamp `json:"timeout"`
	}
)

func (p Params) Clone() Params {
	p.Parts = wallet.CloneAddresses(p.Parts)
	return p
}

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
		Parts:             p.Parts,
		Nonce:             p.Nonce,
		App:               channel.NoApp(),
		LedgerChannel:     true,
	}
}

func (p *Params) UnmarshalJSON(data []byte) error {
	var pj struct {
		ChallengeDuration uint64            `json:"challengeDuration"`
		Parts             []json.RawMessage `json:"parts"`
		Nonce             channel.Nonce     `json:"nonce"`
	}
	if err := json.Unmarshal(data, &pj); err != nil {
		return err
	}

	p.ChallengeDuration = pj.ChallengeDuration
	p.Nonce = pj.Nonce
	p.Parts = make([]wallet.Address, 0, len(pj.Parts))
	for i, rawp := range pj.Parts {
		part := wallet.NewAddress()
		// Hide Address interface to make json.Unmarshaler visible of concrete
		// Address implementation.
		parti := part.(interface{})
		if err := json.Unmarshal(rawp, &parti); err != nil {
			return fmt.Errorf("unmarshaling part[%d]: %w", i, err)
		}
		p.Parts = append(p.Parts, part)
	}
	return nil
}

// CoreState returns the equivalent representation of s as channel.State.
// The returned State is set to have no App, no Data, contains one asset that is
// default initialized and this first assets' balances are set to the Balances
// of s.
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
		Data:    channel.NoData(),
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

func (s State) Sign(acc wallet.Account) (wallet.Sig, error) {
	return channel.Sign(acc, s.CoreState())
}

func VerifySig(signer wallet.Address, state State, sig wallet.Sig) (bool, error) {
	return channel.Verify(signer, state.CoreState(), sig)
}

func (s *StateReg) Clone() *StateReg {
	return &StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone(),
	}
}

func (s *StateReg) Equal(_s StateReg) bool {
	err := s.CoreState().Equal(_s.CoreState())
	return err == nil
}

func (s *StateReg) UnmarshalJSON(data []byte) error {
	sj := struct {
		State   *State          `json:"state"`
		Timeout json.RawMessage `json:"timeout"`
	}{
		State: &s.State,
	}
	if err := json.Unmarshal(data, &sj); err != nil {
		return err
	}

	timeout := NewTimestamp()
	// Hide Timestamp interface to make json.Unmarshaler visible of concrete
	// Timestamp implementation.
	toi := timeout.(interface{})
	if err := json.Unmarshal(sj.Timeout, &toi); err != nil {
		return err
	}
	s.Timeout = timeout
	return nil
}

func (s *StateReg) IsFinalizedAt(ts Timestamp) bool {
	return s.IsFinal || ts.After(s.Timeout)
}

// SignChannel creates signatures on the provided channel state for each
// provided account. The signatures are in the same order as the accounts.
func SignChannel(params Params, state State, accs []wallet.Account) (*SignedChannel, error) {
	sigs := make([]wallet.Sig, 0, len(accs))
	for i, acc := range accs {
		sig, err := state.Sign(acc)
		if err != nil {
			return nil, fmt.Errorf("signing state with account[%d]: %w", i, err)
		}
		sigs = append(sigs, sig)
	}

	return &SignedChannel{
		Params: params,
		State:  state,
		Sigs:   sigs,
	}, nil
}

func (ch *SignedChannel) Clone() *SignedChannel {
	return &SignedChannel{
		Params: ch.Params.Clone(),
		State:  ch.State.Clone(),
		Sigs:   wallet.CloneSigs(ch.Sigs),
	}
}

func ConvertToSignedChannel(req channel.AdjudicatorReq) (*SignedChannel, error) {
	p := req.Params.Clone()
	params := Params{
		ChallengeDuration: p.ChallengeDuration,
		Parts:             p.Parts,
		Nonce:             p.Nonce,
	}

	s := req.Tx.State.Clone()
	if len(s.Balances) != 1 {
		return nil, fmt.Errorf("Only single assets supported.")
	}
	state := State{
		ID:       s.ID,
		Version:  s.Version,
		Balances: s.Balances[0], // We only support a single asset
		IsFinal:  s.IsFinal,
	}

	return &SignedChannel{
		Params: params,
		State:  state,
		Sigs:   req.Tx.Sigs,
	}, nil
}
