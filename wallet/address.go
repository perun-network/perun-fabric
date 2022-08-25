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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"perun.network/go-perun/wallet"
)

// Address identifies a Fabric identity by its X509 certificate's public key.
// The only supported type is *ecdsa.PublicKey because that's the only type used
// in Fabric currently.
type Address ecdsa.PublicKey

// ECDSA returns the public key.
func (a *Address) ECDSA() *ecdsa.PublicKey { return (*ecdsa.PublicKey)(a) }

// Clone duplicates the Address.
func (a *Address) Clone() *Address {
	return &Address{
		X:     new(big.Int).Set(a.X),
		Y:     new(big.Int).Set(a.Y),
		Curve: a.Curve, // curves are global pointers
	}
}

type ecdsaPK struct {
	X     *big.Int
	Y     *big.Int
	Curve string `asn1:"printable"`
}

func (a Address) marshalable() ecdsaPK {
	return ecdsaPK{
		X:     a.X,
		Y:     a.Y,
		Curve: a.Curve.Params().Name,
	}
}

func (a *Address) unmarshal(data ecdsaPK) error {
	a.X = data.X
	a.Y = data.Y
	c, err := ecdsaCurveFromName(data.Curve)
	if err != nil {
		return err
	}
	a.Curve = c
	return nil
}

func (a *Address) MarshalJSON() ([]byte, error) { //nolint:revive
	return json.Marshal(a.marshalable())
}

func (a *Address) UnmarshalJSON(data []byte) error { //nolint:revive
	var pk ecdsaPK
	if err := json.Unmarshal(data, &pk); err != nil {
		return err
	}
	return a.unmarshal(pk)
}

// MarshalBinary allows to marshal Address in ASN1.
func (a *Address) MarshalBinary() ([]byte, error) {
	return asn1.Marshal(a.marshalable())
}

// UnmarshalBinary allows to unmarshal Address from ASN1.
func (a *Address) UnmarshalBinary(data []byte) error {
	var pk ecdsaPK
	if rest, err := asn1.Unmarshal(data, &pk); err != nil {
		return err
	} else if l := len(rest); l > 0 {
		return fmt.Errorf("unexptected rest of length %d during asn1-unmarshaling", l)
	}
	return a.unmarshal(pk)
}

func ecdsaCurveFromName(curve string) (elliptic.Curve, error) {
	switch curve {
	case "P-224":
		return elliptic.P224(), nil
	case "P-256":
		return elliptic.P256(), nil
	case "P-384":
		return elliptic.P384(), nil
	case "P-521":
		return elliptic.P521(), nil
	}
	return nil, fmt.Errorf("unknown curve: %s", curve)
}

func (a *Address) String() string {
	return hex.EncodeToString(elliptic.MarshalCompressed(a.Curve, a.X, a.Y))
}

// Equal returns wether the two addresses are equal. The implementation
// must be equivalent to checking `Address.Cmp(Address) == 0`.
func (a *Address) Equal(other wallet.Address) bool {
	return a.ECDSA().Equal(asECDSA(other))
}

// Cmp checks the ordering of two Addresses according to following definition:
// -1 if (a.X <  b.X) || ((a.X == b.X) && (a.Y < b.Y)).
// 0 if (a.X == b.X) && (a.Y == b.Y).
// +1 if (a.X >  b.X) || ((a.X == b.X) && (a.Y > b.Y)).
// So the X coordinate takes precedence over the Y coordinate.
// Pancis if the passed address is of the wrong type or the curves are not the same.
func (a *Address) Cmp(b wallet.Address) int {
	other := asECDSA(b)
	if a.ECDSA().Curve != other.Curve {
		panic("different ECDSA curves")
	}
	if xCmp := a.X.Cmp(other.X); xCmp != 0 {
		return xCmp
	}
	return a.Y.Cmp(other.Y)
}

func asECDSA(a wallet.Address) *ecdsa.PublicKey {
	return ((a).(*Address)).ECDSA() //nolint:forcetypeassert
}

// NewRandomAddress creates a new Address using the randomness
// provided by rng. The default curve (P-256) is used.
func NewRandomAddress(rng io.Reader) *Address {
	return NewRandomAccount(rng).FabricAddress()
}

// AddressFromX509Certificate extracts the public key from the given certificate.
func AddressFromX509Certificate(cert *x509.Certificate) (*Address, error) {
	pk, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate public key of unexpected type %T", cert.PublicKey)
	}
	return (*Address)(pk), nil
}
