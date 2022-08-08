//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package channel

import (
	"io"

	pchannel "perun.network/go-perun/channel"
)

// asset is the Asset of the connected fabric chain.
// Implements the Perun Asset interface.
// Does not contain any fields since there is only one asset per chain.
type asset struct{}

// Asset is the unique asset that is supported by the chain.
var Asset asset

func (asset) Index() pchannel.Index {
	return 0
}

// Encode does nothing and returns nil since the backend has only one asset.
func (asset) Encode(io.Writer) error {
	return nil
}

// Decode does nothing and returns nil since the backend has only one asset.
func (*asset) Decode(io.Reader) error {
	return nil
}

// MarshalBinary - noop.
func (asset) MarshalBinary() ([]byte, error) { return nil, nil }

// UnmarshalBinary - noop.
func (asset) UnmarshalBinary([]byte) error { return nil }

// Equal returns true if the type matches.
func (asset) Equal(other pchannel.Asset) bool {
	_, ok := other.(asset)
	return ok
}
