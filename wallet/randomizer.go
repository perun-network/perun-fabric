// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
