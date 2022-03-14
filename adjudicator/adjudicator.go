// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
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

func (a *Adjudicator) Deposit(id channel.ID, part wallet.Address, amount *big.Int) error {
	return a.assets.Deposit(id, part, amount)
}

func (a *Adjudicator) Register(ch *SignedChannel) error {
	if err := ValidateChannel(ch); err != nil {
		return err
	}
	id := ch.State.ID

	// Check existing state registration for non-final channels
	if !ch.State.IsFinal {
		if err := a.checkExistingStateReg(ch); err != nil {
			return err
		}
	}

	// check channel funding
	total, err := a.assets.TotalHolding(id, ch.Params.Parts)
	if err != nil {
		return fmt.Errorf("querying total holding: %w", err)
	}
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
		// Update holdings to current state in all other cases so that they can be
		// withdrawn once the channel is finalized.
		if err := a.updateHoldings(ch); err != nil {
			return err
		}
	}

	a.saveStateReg(ch)
	return nil
}

func (a *Adjudicator) checkExistingStateReg(ch *SignedChannel) error {
	reg, ok := a.states[ch.State.ID]
	if !ok {
		return nil
	}

	if now := a.ledger.Now(); now.After(reg.Timeout) {
		return ChallengeTimeoutError{
			Timeout: reg.Timeout,
			Now:     now,
		}
	}

	if ver := ch.State.Version; ver <= reg.Version {
		return VersionError{
			Registered: reg.Version,
			Tried:      ver,
		}
	}

	return nil
}

func (a *Adjudicator) updateHoldings(ch *SignedChannel) error {
	for i, part := range ch.Params.Parts {
		if err := a.assets.SetHolding(ch.State.ID, part, ch.State.Balances[i]); err != nil {
			return fmt.Errorf("updating holding[%d]: %w", i, err)
		}
	}
	return nil
}

func (a *Adjudicator) saveStateReg(ch *SignedChannel) {
	// determine timeout by channel finality
	to := a.ledger.Now()
	if !ch.State.IsFinal {
		to = to.Add(ch.Params.ChallengeDuration)
	}

	// save state to states registry
	a.states[ch.State.ID] = &StateReg{
		State:   ch.State,
		Timeout: to,
	}
}

// Withdraw withdraws all funds of participant part in the finalized channel id
// to themself via the AssetHolder. It returns the withdrawn amount.
func (a *Adjudicator) Withdraw(id channel.ID, part wallet.Address) (*big.Int, error) {
	if reg, ok := a.states[id]; !ok {
		return nil, ErrUnknownChannel
	} else if now := a.ledger.Now(); !reg.IsFinalizedAt(now) {
		return nil, ChallengeTimeoutError{
			Timeout: reg.Timeout,
			Now:     now,
		}
	}

	return a.assets.Withdraw(id, part)
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
		if ok, err := channel.Verify(ch.Params.Parts[i], ch.State.CoreState(), sig); err != nil {
			return ValidationError{fmt.Errorf("validating sig[%d]: %w", i, err)}
		} else if !ok {
			return ValidationError{fmt.Errorf("sig[%d] invalid", i)}
		}
	}

	return nil
}
