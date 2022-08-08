package adjudicator

import (
	"math/big"
)

// Asset is utilized for constructing asset types to operate Perun state channels on.
type Asset interface {
	// Mint creates the desired amount of token for the given id.
	// Note that id must be authenticated first.
	Mint(id string, amount *big.Int) error

	// Burn removes the desired amount of token from the given id.
	// Note that id must be authenticated first.
	Burn(id string, amount *big.Int) error

	// Transfer sends the desired amount of tokens from sender to receiver.
	// Note that sender must be authenticated first.
	Transfer(sender string, receiver string, amount *big.Int) error

	// BalanceOf returns the amount of tokens the given id holds.
	BalanceOf(id string) (*big.Int, error)
}
