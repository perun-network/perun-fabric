// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"fmt"
	"math/rand"

	"perun.network/go-perun/wallet"
)

type Wallet map[wallet.AddrKey]*Account

// NewWallet creates a new Wallet that will contain the optionally provided
// accounts.
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

func (w Wallet) Add(acc *Account) {
	w[wallet.Key(acc.Address())] = acc
}

func (w Wallet) NewRandomAccount(rng *rand.Rand) wallet.Account {
	acc := NewRandomAccount(rng)
	w.Add(acc)
	return acc
}

// LockAll - noop
func (w Wallet) LockAll() {}

// IncrementUsage - noop
func (w Wallet) IncrementUsage(wallet.Address) {}

// DecrementUsage - noop
func (w Wallet) DecrementUsage(wallet.Address) {}
