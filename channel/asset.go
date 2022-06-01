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

// MarshalBinary - noop
func (asset) MarshalBinary() ([]byte, error) { return nil, nil }

// UnmarshalBinary - noop
func (asset) UnmarshalBinary([]byte) error { return nil }

// Equal returns true
func (asset) Equal(other pchannel.Asset) bool {
	return true
}
