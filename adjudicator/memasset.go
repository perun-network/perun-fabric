package adjudicator

import (
	"fmt"
	"math/big"
)

// MemAsset is an in-memory asset for testing.
// As it is for testing there is no central banker.
type MemAsset struct {
	holdings map[string]*big.Int
}

// NewMemAsset generates a new in-memory Asset.
func NewMemAsset() *MemAsset {
	return &MemAsset{
		holdings: make(map[string]*big.Int),
	}
}

// Mint creates the desired amount of token for the given id.
func (m MemAsset) Mint(id AccountID, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	current, _ := m.BalanceOf(id) // No error expected
	current.Add(current, amount)
	m.holdings[string(id)] = new(big.Int).Set(current)
	return nil
}

// Burn removes the desired amount of token from the given id.
func (m MemAsset) Burn(id AccountID, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot burn zero/negative amount")
	}

	// Get current balance.
	current, _ := m.BalanceOf(id) // No error expected.
	current.Sub(current, amount)

	if current.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("not enought funds to burn the requested amount")
	}

	m.holdings[string(id)] = new(big.Int).Set(current)
	return nil
}

// Transfer checks if the proposed transfer is valid and
// transfers the given amount of coins from the sender to the receiver.
func (m MemAsset) Transfer(sender AccountID, receiver AccountID, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("cannot transfer negative amount")
	}

	// Check balance of sender.
	senderBal, _ := m.BalanceOf(sender) // No error expected.
	if !(senderBal.Cmp(amount) >= 0) {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, _ := m.BalanceOf(receiver) // No error expected.

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	m.holdings[string(sender)] = new(big.Int).Set(senderBal)
	m.holdings[string(receiver)] = new(big.Int).Set(receiverBal)
	return nil
}

// BalanceOf returns the amount of tokens the given id holds.
// If the id is unknown, zero is returned.
func (m MemAsset) BalanceOf(id AccountID) (*big.Int, error) {
	current, ok := m.holdings[string(id)]
	if !ok {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(current), nil
}
