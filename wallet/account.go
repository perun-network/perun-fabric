// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"

	"perun.network/go-perun/wallet"
)

var defaultCurve = elliptic.P256()

// Account identifies a Fabric identity by its X509 certificate's private key.
// The only supported type is *ecdsa.PrivateKey because that's the only type used
// in Fabric currently.
type Account ecdsa.PrivateKey

// ECDSA returns the private key.
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

// Address of this Account.
func (a *Account) Address() wallet.Address {
	return a.FabricAddress()
}

// FabricAddress returns the public key of this account as this package's
// Address type.
func (a *Account) FabricAddress() *Address {
	return (*Address)(&a.PublicKey)
}

// SignData signs the data with this account. The data is hashed before signing.
func (a *Account) SignData(data []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, a.ECDSA(), Hash(data))
	if err != nil {
		return nil, fmt.Errorf("ecdsa.Sign: %w", err)
	}
	return marshalSig(a.Curve, r, s), nil
}

// Hash returns the SHA256 of msg.
func Hash(msg []byte) []byte {
	h := sha256.Sum256(msg)
	return h[:]
}

func marshalSig(curve elliptic.Curve, r, s *big.Int) wallet.Sig {
	ps := pointByteSize(curve)
	sig := make([]byte, sigSize(ps))
	sig[0] = ps
	r.FillBytes(sig[1 : ps+1])
	s.FillBytes(sig[ps+1 : 2*ps+1])
	return sig
}

func unmarshalSig(sig []byte) (*big.Int, *big.Int, error) {
	if len(sig) == 0 {
		return nil, nil, errors.New("UnmarshalSig: empty signature")
	}
	ps, r, s := sig[0], new(big.Int), new(big.Int)
	if l, ss := len(sig), sigSize(ps); l != ss {
		return nil, nil, fmt.Errorf("UnmarshalSig: signature has wrong length %d, expected %d", l, ss)
	}
	r.SetBytes(sig[1 : ps+1])
	s.SetBytes(sig[ps+1 : 2*ps+1])
	return r, s, nil
}

// DecodeSig decodes the given io.Reader to a wallet signature.
func DecodeSig(r io.Reader) (wallet.Sig, error) {
	sig := make([]byte, 1)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, fmt.Errorf("reading sig size: %w", err)
	}
	sig = append(sig, make([]byte, sig[0]*2)...) //nolint:makezero,gomnd
	_, err := io.ReadFull(r, sig[1:])
	return sig, err
}

func pointByteSize(curve elliptic.Curve) byte {
	// rounding up is necessary for P521
	return byte((curve.Params().BitSize + 7) / 8) //nolint:gomnd
}

func sigSize(pointByteSize byte) int {
	return int(pointByteSize)*2 + 1
}
