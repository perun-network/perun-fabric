// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"fmt"
	"math/rand"

	"perun.network/go-perun/wallet"
)

// Wallet contains multiple Account's identifiable by their AddrKey.
type Wallet map[wallet.AddrKey]*Account

// NewWallet creates a new Wallet that will contain the optionally provided accounts.
func NewWallet(accs ...*Account) Wallet {
	w := make(Wallet, len(accs))
	for _, acc := range accs {
		w.Add(acc)
	}
	return w
}

// Unlock returns the account for the provided address.
func (w Wallet) Unlock(a wallet.Address) (wallet.Account, error) {
	acc, ok := w[wallet.Key(a)]
	if !ok {
		return nil, fmt.Errorf("unknown address: %s", a)
	}
	return acc, nil
}

// Add adds the given Account to the Wallet.
func (w Wallet) Add(acc *Account) {
	w[wallet.Key(acc.Address())] = acc
}

// NewRandomAccount generates a new random Account and adds it to the Wallet.
func (w Wallet) NewRandomAccount(rng *rand.Rand) wallet.Account {
	acc := NewRandomAccount(rng)
	w.Add(acc)
	return acc
}

// LockAll - noop.
func (w Wallet) LockAll() {}

// IncrementUsage - noop.
func (w Wallet) IncrementUsage(wallet.Address) {}

// DecrementUsage - noop.
func (w Wallet) DecrementUsage(wallet.Address) {}
