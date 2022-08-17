// SPDX-License-Identifier: Apache-2.0

package adjudicator

import (
	"errors"
	"fmt"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// Adjudicator is an abstract implementation of the adjudicator smart contract.
type Adjudicator struct {
	ledger     Ledger       // ledger stores channel states.
	holdings   *AssetHolder // holdings manages per channel holdings.
	asset      Asset        // asset is the Asset the channels use. It can also be used independently of the channels.
	identifier AccountID    // identifier is the chaincode id for sending and receiving funds on.
}

// NewAdjudicator generates a new Adjudicator with an identifier, holding ledger and asset ledger.
func NewAdjudicator(id string, ledger Ledger, asset Asset) *Adjudicator {
	return &Adjudicator{
		ledger:     ledger,
		holdings:   NewAssetHolder(ledger),
		asset:      asset,
		identifier: AccountID(id),
	}
}

// Register verifies the given SignedChannel, updates the holdings and saves a new StateReg.
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
	if !ch.State.IsFinal {
		to = to.Add(ch.Params.ChallengeDuration)
	}

	// save StateReg to ledger
	return a.ledger.PutState(&StateReg{
		State:   ch.State,
		Timeout: to,
	})
}

// StateReg fetches the current state from the ledger and returns it.
// If no state is found under the given channel.ID an error is returned.
func (a *Adjudicator) StateReg(id channel.ID) (*StateReg, error) {
	reg, err := a.ledger.GetState(id)
	if IsNotFoundError(err) {
		return nil, ErrUnknownChannel
	} else if err != nil {
		return nil, fmt.Errorf("querying ledger: %w", err)
	}
	return reg, nil
}

// Withdraw withdraws all funds of participant Part in the finalized channel id
// to the given Receiver. It returns the withdrawn amount.
func (a *Adjudicator) Withdraw(swr SignedWithdrawReq) (*big.Int, error) {
	if reg, err := a.StateReg(swr.Req.ID); err != nil {
		return nil, err
	} else if now := a.ledger.Now(); !reg.IsFinalizedAt(now) {
		return nil, ChallengeTimeoutError{
			Timeout: reg.Timeout,
			Now:     now,
		}
	}

	// Verify signature.
	sigValid, err := swr.Verify(swr.Req.Part)
	if err != nil {
		return nil, err
	}
	if !sigValid {
		return nil, fmt.Errorf("withdraw request signature invalid")
	}

	// Withdraw from channel.
	holding, err := a.holdings.Withdraw(swr.Req.ID, swr.Req.Part)
	if err != nil {
		return nil, err
	}

	// Send funds back.
	err = a.asset.Transfer(a.identifier, swr.Req.Receiver, holding)
	if err != nil {
		return nil, err
	}
	return holding, nil
}

// ValidateChannel checks if the given parameters in SignedChannel are in itself consistent.
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

// Deposit transfers the given amount of coins from the callee to the channel with the specified channel ID.
// The funds are stored in the channel under the participant's wallet address.
func (a *Adjudicator) Deposit(callee AccountID, chID channel.ID, part wallet.Address, amount *big.Int) error {
	// Transfer funds to channel.
	err := a.asset.Transfer(callee, a.identifier, amount)
	if err != nil {
		return err
	}

	// Register deposit.
	return a.holdings.Deposit(chID, part, amount)
}

// Holding returns the current holding amount of the given participant in the channel.
func (a *Adjudicator) Holding(id channel.ID, part wallet.Address) (*big.Int, error) {
	return a.holdings.Holding(id, part)
}

// TotalHolding returns the sum of all participant holdings in the channel.
func (a *Adjudicator) TotalHolding(id channel.ID, parts []wallet.Address) (*big.Int, error) {
	return a.holdings.TotalHolding(id, parts)
}

// Mint generates the given amount of asset tokens for the callee.
func (a *Adjudicator) Mint(callee AccountID, amount *big.Int) error {
	return a.asset.Mint(callee, amount)
}

// Burn destroys the given amount of asset tokens for the callee.
func (a *Adjudicator) Burn(callee AccountID, amount *big.Int) error {
	return a.asset.Burn(callee, amount)
}

// Transfer sends the given amount of asset tokens from the sender to the receiver.
func (a *Adjudicator) Transfer(sender AccountID, receiver AccountID, amount *big.Int) error {
	return a.asset.Transfer(sender, receiver, amount)
}

// BalanceOfID returns the asset token balance of the given user identifier.
func (a *Adjudicator) BalanceOfID(id AccountID) (*big.Int, error) {
	return a.asset.BalanceOf(id)
}
