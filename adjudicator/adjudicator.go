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
	ledger     Ledger
	asset      Asset
	holdings   *AssetHolder
	identifier string
}

func NewAdjudicator(id string, ledger Ledger, asset Asset) *Adjudicator {
	return &Adjudicator{
		ledger:     ledger,
		asset:      asset,
		holdings:   NewAssetHolder(ledger),
		identifier: id,
	}
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
	total, err := a.holdings.TotalHolding(id, ch.Params.Parts)
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

	return a.saveStateReg(ch)
}

func (a *Adjudicator) checkExistingStateReg(ch *SignedChannel) error {
	reg, err := a.ledger.GetState(ch.State.ID)
	if IsNotFoundError(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("querying ledger: %w", err)
	}

	if now := a.ledger.Now(); now.After(reg.Timeout) {
		return ChallengeTimeoutError{
			Timeout: reg.Timeout,
			Now:     now,
		}
	}

	// allow registration of same version for idempotence of Register
	if ver := ch.State.Version; ver < reg.Version {
		return VersionError{
			Registered: reg.Version,
			Tried:      ver,
		}
	}

	return nil
}

func (a *Adjudicator) updateHoldings(ch *SignedChannel) error {
	for i, part := range ch.Params.Parts {
		if err := a.holdings.SetHolding(ch.State.ID, part, ch.State.Balances[i]); err != nil {
			return fmt.Errorf("updating holding[%d]: %w", i, err)
		}
	}
	return nil
}

func (a *Adjudicator) saveStateReg(ch *SignedChannel) error {
	// determine timeout by channel finality
	to := a.ledger.Now()
	ch.State.Now = to

	if !ch.State.IsFinal {
		to = to.Add(ch.Params.ChallengeDuration)
	}

	// save StateReg to ledger
	return a.ledger.PutState(&StateReg{
		State:   ch.State,
		Timeout: to,
	})
}

func (a *Adjudicator) StateReg(id channel.ID) (*StateReg, error) {
	reg, err := a.ledger.GetState(id)
	if IsNotFoundError(err) {
		return nil, ErrUnknownChannel
	} else if err != nil {
		return nil, fmt.Errorf("querying ledger: %w", err)
	}

	reg.State.Now = a.ledger.Now() // Add the ledger timestamp.
	return reg, nil
}

// Withdraw withdraws all funds of participant part in the finalized channel id
// to themself via the AssetHolder. It returns the withdrawn amount.
func (a *Adjudicator) Withdraw(callee string, id channel.ID, part wallet.Address) (*big.Int, error) { // TODO: Remove callee for signed request.
	if reg, err := a.StateReg(id); err != nil {
		return nil, err
	} else if now := a.ledger.Now(); !reg.IsFinalizedAt(now) {
		return nil, ChallengeTimeoutError{
			Timeout: reg.Timeout,
			Now:     now,
		}
	}

	// Withdraw from channel.
	holding, err := a.holdings.Withdraw(id, part)
	if err != nil {
		return nil, err
	}

	// Send funds back to participant.
	// TODO: Get ID to send funds back to from signed request.
	// TODO: Check if signature is valid in regard to the participant address!
	err = a.asset.Transfer(a.identifier, callee, holding)
	if err != nil {
		return nil, err
	}
	return holding, nil
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
		if ok, err := VerifySig(ch.Params.Parts[i], ch.State, sig); err != nil {
			return ValidationError{fmt.Errorf("validating sig[%d]: %w", i, err)}
		} else if !ok {
			return ValidationError{fmt.Errorf("sig[%d] invalid", i)}
		}
	}

	return nil
}

func (a *Adjudicator) Deposit(calleeID string, chID channel.ID, part wallet.Address, amount *big.Int) error {
	// Transfer funds to channel.
	err := a.asset.Transfer(calleeID, a.identifier, amount)
	if err != nil {
		return err
	}

	// Register deposit.
	err = a.holdings.Deposit(chID, part, amount)
	if err != nil {
		// Ensure funds not stuck in channel if deposit fails.
		transferErr := a.asset.Transfer(a.identifier, calleeID, amount)
		if err != nil {
			return transferErr
		}
	}
	return err
}

func (a *Adjudicator) Holding(id channel.ID, part wallet.Address) (*big.Int, error) {
	return a.holdings.Holding(id, part)
}

func (a *Adjudicator) TotalHolding(id channel.ID, parts []wallet.Address) (*big.Int, error) {
	return a.holdings.TotalHolding(id, parts)
}

func (a *Adjudicator) Mint(calleeID string, amount *big.Int) error {
	return a.asset.Mint(calleeID, amount)
}

func (a *Adjudicator) Burn(calleeID string, amount *big.Int) error {
	return a.asset.Burn(calleeID, amount)
}

func (a *Adjudicator) Transfer(senderID string, receiverID string, amount *big.Int) error {
	return a.asset.Transfer(senderID, receiverID, amount)
}

func (a *Adjudicator) BalanceOfID(userID string) (*big.Int, error) {
	return a.asset.BalanceOf(userID)
}
