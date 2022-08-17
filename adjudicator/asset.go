package adjudicator

import (
	"math/big"
)

// AccountID represents the ID of a client.
// It is used for minting, burning, receiving and sending funds.
// Ensure it is unique for every client interacting with Asset and no impersonation is possible.
type AccountID string

// Asset is a basic interface for creating tokens with.
type Asset interface {
	// Mint creates the desired amount of token for the given id.
	// Note that id must be authenticated first.
	Mint(id AccountID, amount *big.Int) error

	// Burn removes the desired amount of token from the given id.
	// Note that id must be authenticated first.
	Burn(id AccountID, amount *big.Int) error

	// Transfer sends the desired amount of tokens from sender to receiver.
	// Note that sender must be authenticated first.
	Transfer(sender AccountID, receiver AccountID, amount *big.Int) error

	// BalanceOf returns the amount of tokens the given id holds.
	BalanceOf(id AccountID) (*big.Int, error)
}
