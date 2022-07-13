package test

import (
	fwallet "github.com/perun-network/perun-fabric/wallet"
	"math/rand"
	"perun.network/go-perun/wallet"
)

type ConstWallet struct {
	mainWallet      fwallet.Wallet
	constAccAddress wallet.Address
	usage           int
}

// NewConstTestWallet creates a new wallet with special behavior.
// It generates one account during initialization and provides this constantly via
// GetConstAccount - and on the first (repeat) calls of NewRandomAccount.
func NewConstTestWallet(rng *rand.Rand, repeat int) *ConstWallet {
	w := fwallet.NewWallet()
	acc := w.NewRandomAccount(rng)
	addr := acc.Address()
	return &ConstWallet{w, addr, repeat}
}

// Unlock returns the account for the provided address.
func (w ConstWallet) Unlock(a wallet.Address) (wallet.Account, error) {
	return w.mainWallet.Unlock(a)
}

// NewRandomAccount returns the constant account generated during initialization on its first invocation.
// After this, it behaves as expected and generates random accounts which are added to the wallet.
func (w ConstWallet) NewRandomAccount(rng *rand.Rand) wallet.Account {
	if w.usage > 0 {
		w.usage -= 1
		return w.GetConstAccount()
	}
	return w.mainWallet.NewRandomAccount(rng)
}

// GetConstAccount returns the constant account.
func (w ConstWallet) GetConstAccount() wallet.Account {
	acc, _ := w.mainWallet.Unlock(w.constAccAddress)
	return acc
}

// LockAll - noop
func (w ConstWallet) LockAll() {}

// IncrementUsage - noop
func (w ConstWallet) IncrementUsage(wallet.Address) {}

// DecrementUsage - noop
func (w ConstWallet) DecrementUsage(wallet.Address) {}
