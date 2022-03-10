// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
)

// Adjudicator is an abstract implementation of the adjudicator smart
// contract.
type Adjudicator struct {
	ledger Ledger
	assets *AssetHolder
	states map[channel.ID]*StateReg
}

func NewAdjudicator(ledger Ledger) *Adjudicator {
	return &Adjudicator{
		ledger: ledger,
		assets: NewAssetHolder(ledger),
		states: make(map[channel.ID]*StateReg),
	}
}

func (a *Adjudicator) Deposit(id channel.ID, part Address, amount *big.Int) error {
	return a.assets.Deposit(id, &part, amount)
}

func (a *Adjudicator) Register(ch *SignedChannel) error {
	if err := ValidateChannel(ch); err != nil {
		return err
	}
	id := ch.State.ID

	// TODO:
	// - Check for existing registered channel
	//   - Check its timeout
	//   - Check version increase

	total, err := a.assets.TotalHolding(id, AsWalletAddresses(ch.Params.Parts))
	if err != nil {
		return fmt.Errorf("querying total holding: %w", err)
	}

	// check underfunded channel
	if chTotal := ch.State.Total(); total.Cmp(chTotal) == -1 {
		// allow version 0 underfunded channels for funds recovery
		if ch.State.Version != 0 {
			return &UnderfundedError{
				Version: ch.State.Version,
				Total:   chTotal,
				Funded:  total,
			}
		}
	} else {
		if err := a.updateHoldings(ch); err != nil {
			return err
		}
	}

	// determine timeout by channel finality
	to := a.ledger.Now().(*StdTimestamp)
	if ch.State.IsFinal {
		to = to.Add(ch.Params.ChallengeDuration).(*RegTimestamp)
	}

	// save state to states registry
	a.states[id] = &StateReg{
		State:   ch.State,
		Timeout: to,
	}
	return nil
}

func (a *Adjudicator) updateHoldings(ch *SignedChannel) error {
	for i, part := range ch.Params.Parts {
		if err := a.assets.SetHolding(ch.State.ID, &part, ch.State.Balances[i]); err != nil {
			return fmt.Errorf("updating holding[%d]: %w", i, err)
		}
	}
	return nil
}

func ValidateChannel(ch *SignedChannel) error {
	if ch.Params.ID() != ch.State.ID {
		return ValidationError{errors.New("channel id mismatch")}
	}

	n := len(ch.Params.Parts)
	if n != len(ch.State.Balances) {
		return ValidationError{errors.New("balances dimension mismatch")}
	}
	if n != len(ch.Sigs) {
		return ValidationError{errors.New("sigs dimension mismatch")}
	}

	for i, sig := range ch.Sigs {
		if ok, err := channel.Verify(&ch.Params.Parts[i], ch.State.CoreState(), sig); err != nil {
			return ValidationError{fmt.Errorf("validating sig[%d]: %w", i, err)}
		} else if !ok {
			return ValidationError{fmt.Errorf("sig[%d] invalid", i)}
		}
	}

	return nil
}

type ValidationError struct{ error }

func (ve ValidationError) Unwrap() error {
	return ve.error
}

type UnderfundedError struct {
	Version uint64
	Total   *big.Int
	Funded  *big.Int
}

func (ue UnderfundedError) Error() string {
	return fmt.Sprintf("channel underfunded (%v < %v, version %d)", ue.Funded, ue.Total, ue.Version)
}
