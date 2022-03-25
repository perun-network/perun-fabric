// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type Backend struct{}

// CalcID calculates the channel id of a channel from its parameters. Usually,
// this should be a hash digest of some or all fields of the parameters.
// In order to guarantee non-malleability of States, any parameters omitted
// from the CalcID digest need to be signed together with the State in
// Sign().
func (Backend) CalcID(params *channel.Params) (id channel.ID) {
	hash := sha256.New()
	if err := params.Encode(hash); err != nil {
		panic("error encoding Params: " + err.Error())
	}
	copy(id[:], hash.Sum(nil))
	return id
}

// Sign signs a channel's State with the given Account.
// Returns the signature or an error.
func (Backend) Sign(acc wallet.Account, state *channel.State) (wallet.Sig, error) {
	var buf bytes.Buffer
	if err := state.Encode(&buf); err != nil {
		return nil, fmt.Errorf("encoding State: %w", err)
	}
	return acc.SignData(buf.Bytes())
}

// Verify verifies that the provided signature on the state belongs to the
// provided address.
func (Backend) Verify(addr wallet.Address, state *channel.State, sig wallet.Sig) (bool, error) {
	var buf bytes.Buffer
	if err := state.Encode(&buf); err != nil {
		return false, fmt.Errorf("encoding State: %w", err)
	}
	return wallet.VerifySignature(buf.Bytes(), sig, addr)
}

// NewAsset returns a variable of type Asset, which can be used
// for unmarshalling an asset from its binary representation.
func (Backend) NewAsset() channel.Asset {
	return noAsset{}
}

func (b Backend) NewRandomAsset(*rand.Rand) channel.Asset {
	return noAsset{}
}

type noAsset struct{}

// MarshalBinary - noop
func (noAsset) MarshalBinary() ([]byte, error) { return nil, nil }

// UnmarshalBinary - noop
func (noAsset) UnmarshalBinary([]byte) error { return nil }

// Equal returns true iff the other asset is also a noAsset
func (noAsset) Equal(other channel.Asset) bool {
	_, ok := other.(noAsset)
	return ok
}