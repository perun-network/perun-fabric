// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"perun.network/go-perun/wallet"
)

type Account ecdsa.PrivateKey

func (a *Account) ECDSA() *ecdsa.PrivateKey { return (*ecdsa.PrivateKey)(a) }

// NewRandomAccount creates a new Account using the randomness
// provided by rng. The default curve (P-256) is used.
func NewRandomAccount(rng io.Reader) *Account {
	sk, err := ecdsa.GenerateKey(defaultCurve, rng)
	if err != nil {
		panic("error generating ECDSA secret key: " + err.Error())
	}
	return (*Account)(sk)
}

// Address belonging to this Account.
func (a *Account) Address() wallet.Address {
	return a.ECDSAAddress()
}

// ECDSAAddress returns the public key of this account as this package's Address
// type.
func (a *Account) ECDSAAddress() *Address {
	return (*Address)(&a.PublicKey)
}

// SignData signs the data with this account. The data is hashed before signing.
func (a *Account) SignData(data []byte) ([]byte, error) {
	return ecdsa.SignASN1(rand.Reader, a.ECDSA(), Hash(data))
}

// Hash returns the SHA256 of msg.
func Hash(msg []byte) []byte {
	h := sha256.Sum256(msg)
	return h[:]
}
