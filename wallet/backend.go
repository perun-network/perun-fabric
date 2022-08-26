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
