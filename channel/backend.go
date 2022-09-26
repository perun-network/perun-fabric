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

package channel

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// Backend provides basic functionalities for fabric.
type Backend struct{}

// CalcID calculates the channel id of a channel from its parameters. Usually,
// this should be a hash digest of some or all fields of the parameters.
// In order to guarantee non-malleability of States, any parameters omitted
// from the CalcID digest need to be signed together with the State in
// Sign().
func (Backend) CalcID(params *channel.Params) channel.ID {
	hash := sha256.New()
	if err := params.Encode(hash); err != nil {
		panic("error encoding Params: " + err.Error())
	}
	id := channel.ID{}
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

// NewAsset returns a new fabric asset.
func (Backend) NewAsset() channel.Asset {
	return asset{}
}

// NewRandomAsset returns a new fabric asset.
func (b Backend) NewRandomAsset(*rand.Rand) channel.Asset {
	return asset{}
}
