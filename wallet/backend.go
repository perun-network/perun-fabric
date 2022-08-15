// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"io"

	"perun.network/go-perun/wallet"
)

// Backend provides useful methods for hl fabric.
type Backend struct{}

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func (b Backend) NewAddress() wallet.Address {
	return new(Address)
}

// DecodeSig reads a signature from the provided reader.
func (b Backend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	return DecodeSig(r)
}

// VerifySignature verifies that signature sig is a valid signature by a on
// message msg.
// If the signature does not match the address, it returns false, nil.
func (b Backend) VerifySignature(msg []byte, sig wallet.Sig, a wallet.Address) (bool, error) {
	r, s, err := unmarshalSig(sig)
	if err != nil {
		return false, fmt.Errorf("unmarshaling sig: %w", err)
	}
	return ecdsa.Verify(asECDSA(a), Hash(msg), r, s), nil
}
