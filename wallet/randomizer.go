package wallet

import (
	"math/rand"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

// Randomizer wraps a Wallet to derive new random parameters from it.
type Randomizer struct {
	wallet Wallet
}

// NewRandomizer returns a Randomizer containing a fresh, random Wallet.
func NewRandomizer() *Randomizer {
	return &Randomizer{
		wallet: NewWallet(),
	}
}

// NewRandomAddress should return a new random address generated from the
// passed rng.
func (r *Randomizer) NewRandomAddress(rng *rand.Rand) wallet.Address {
	return NewRandomAddress(rng)
}

// RandomWallet returns a fixed random wallet that is part of the
// randomizer's state.
func (r *Randomizer) RandomWallet() test.Wallet { return r.wallet }

// NewWallet returns a fresh, temporary Wallet that doesn't hold any
// accounts yet.
func (r *Randomizer) NewWallet() test.Wallet { return NewWallet() }
