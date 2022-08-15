package adjudicator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"math/big"
	"perun.network/go-perun/wire/perunio"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type (
	// Params are the parameters of a state channel.
	Params struct {
		ChallengeDuration uint64           `json:"challengeDuration"`
		Parts             []wallet.Address `json:"parts"`
		Nonce             channel.Nonce    `json:"nonce"`
	}

	// State is a state of a state channel.
	State struct {
		ID       channel.ID    `json:"id"`
		Version  uint64        `json:"version"`
		Balances []channel.Bal `json:"balances"`
		IsFinal  bool          `json:"final"`
	}

	// SignedChannel contains signatures on Params and State and is used for registering new states.
	SignedChannel struct {
		Params Params       `json:"params"`
		State  State        `json:"state"`
		Sigs   []wallet.Sig `json:"sigs"`
	}

	// WithdrawReq are parameters needed to withdraw funds from a state channel.
	WithdrawReq struct {
		ID       channel.ID     `json:"id"`
		Part     wallet.Address `json:"part"`
		Receiver string         `json:"receiver"`
	}

	// SignedWithdrawReq contains a signature over a WithdrawReq to check its validity.
	SignedWithdrawReq struct {
		Req WithdrawReq `json:"req"`
		Sig wallet.Sig  `json:"sig"`
	}

	// StateReg adds a Timeout to the State to indicate the states challenge timeout.
	StateReg struct {
		State   `json:"state"`
		Timeout Timestamp `json:"timeout"`
	}
)

// Clone duplicates the params.
func (p Params) Clone() Params {
	p.Parts = wallet.CloneAddresses(p.Parts)
	return p
}

// ID return the params channel id.
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

// UnmarshalJSON implements custom unmarshalling for Params.
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
		parti := part.(interface{}) //nolint:forcetypeassert
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
		IsFinal: s.IsFinal,
		App:     channel.NoApp(),
		Data:    channel.NoData(),
		Allocation: channel.Allocation{
			Assets:   []channel.Asset{channel.NewAsset()},
			Balances: channel.Balances{s.Balances},
		},
	}
}

// Total returns the total balance of the State.
func (s State) Total() channel.Bal {
	total := new(big.Int)
	for _, bal := range s.Balances {
		total.Add(total, bal)
	}
	return total
}

// Clone duplicates the State.
func (s State) Clone() State {
	bals := channel.CloneBals(s.Balances)
	s.Balances = bals
	// Other fields are value types, so done
	return s
}

// Sign signs the State with a given account.
func (s State) Sign(acc wallet.Account) (wallet.Sig, error) {
	return channel.Sign(acc, s.CoreState())
}

// VerifySig verifies the signature on a State.
func VerifySig(signer wallet.Address, state State, sig wallet.Sig) (bool, error) {
	return channel.Verify(signer, state.CoreState(), sig)
}

// Clone duplicates the StateReg.
func (s *StateReg) Clone() *StateReg {
	return &StateReg{
		State:   s.State.Clone(),
		Timeout: s.Timeout.Clone(),
	}
}

// Equal checks if the given StateReg is equal.
func (s *StateReg) Equal(sr StateReg) bool {
	err := s.CoreState().Equal(sr.CoreState())
	return err == nil && s.Timeout.Equal(sr.Timeout)
}

// IsFinalizedAt checks if the registered state is final.
// This is the case if either the isFinal flag is true or the timeout passed.
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

// Clone duplicates a SignedChannel.
func (ch *SignedChannel) Clone() *SignedChannel {
	return &SignedChannel{
		Params: ch.Params.Clone(),
		State:  ch.State.Clone(),
		Sigs:   wallet.CloneSigs(ch.Sigs),
	}
}

// ConvertToSignedChannel takes a AdjudicatorReq and generates a SignedChannel from it.
func ConvertToSignedChannel(req channel.AdjudicatorReq) (*SignedChannel, error) {
	p := req.Params.Clone()
	params := Params{
		ChallengeDuration: p.ChallengeDuration,
		Parts:             p.Parts,
		Nonce:             p.Nonce,
	}

	s := req.Tx.State.Clone()
	if len(s.Balances) != 1 {
		return nil, fmt.Errorf("only single assets supported")
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

// SignWithdrawRequest generates a WithdrawReq and signs it with the given account to return a SignedWithdrawReq.
func SignWithdrawRequest(acc wallet.Account, channel channel.ID, receiver string) (*SignedWithdrawReq, error) {
	req := WithdrawReq{
		ID:       channel,
		Part:     acc.Address(),
		Receiver: receiver,
	}

	sig, err := req.Sign(acc)
	if err != nil {
		return nil, err
	}

	return &SignedWithdrawReq{
		Req: req,
		Sig: sig,
	}, nil
}

// UnmarshalJSON implements custom unmarshalling for WithdrawReq to deal with the participant address.
func (wr *WithdrawReq) UnmarshalJSON(data []byte) error {
	var wrj struct {
		ID       channel.ID      `json:"id"`
		Part     json.RawMessage `json:"part"`
		Receiver string          `json:"receiver"`
	}
	if err := json.Unmarshal(data, &wrj); err != nil {
		return err
	}

	wr.ID = wrj.ID
	wr.Receiver = wrj.Receiver

	part := wallet.NewAddress()
	parti := part.(interface{}) //nolint:forcetypeassert
	if err := json.Unmarshal(wrj.Part, &parti); err != nil {
		return fmt.Errorf("unmarshaling part: %w", err)
	}
	wr.Part = part
	return nil
}

// Sign signs a withdraw request with the given Account.
// Returns the signature or an error.
func (wr WithdrawReq) Sign(acc wallet.Account) (wallet.Sig, error) {
	var buf bytes.Buffer
	if err := wr.Encode(&buf); err != nil {
		return nil, fmt.Errorf("encoding WithdrawReq: %w", err)
	}
	return acc.SignData(buf.Bytes())
}

// Verify verifies that the provided signature on the signed withdraw request belongs to the
// provided address.
func (swr SignedWithdrawReq) Verify(addr wallet.Address) (bool, error) {
	var buf bytes.Buffer
	if err := swr.Req.Encode(&buf); err != nil {
		return false, fmt.Errorf("encoding WithdrawReq: %w", err)
	}
	return wallet.VerifySignature(buf.Bytes(), swr.Sig, addr)
}

// Encode encodes a withdraw request into an `io.Writer` or returns an `error`.
func (wr WithdrawReq) Encode(w io.Writer) error {
	return errors.WithMessage(
		perunio.Encode(w, wr.ID, wr.Part, wr.Receiver), "WithdrawReq encode")
}

// Decode decodes a withdraw request from an `io.Reader` or returns an `error`.
func (wr WithdrawReq) Decode(r io.Reader) error {
	return errors.WithMessage(
		perunio.Decode(r, &wr.ID, &wr.Part, &wr.Receiver), "WithdrawReq decode")
}
