package adjudicator

import (
	"fmt"
	"math/big"
)

type MemAsset struct {
	holdings map[string]*big.Int
}

func (m MemAsset) Mint(id string, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	current, _ := m.BalanceOf(id) // No error expected
	current.Add(current, amount)
	m.holdings[id] = new(big.Int).Set(current)
	return nil
}

func (m MemAsset) Burn(id string, amount *big.Int) error {
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

	m.holdings[id] = new(big.Int).Set(current)
	return nil
}

func (m MemAsset) Transfer(id string, receiver string, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("cannot transfer negative amount")
	}

	// Check balance of sender.
	senderBal, _ := m.BalanceOf(id) // No error expected.
	if !(senderBal.Cmp(amount) >= 0) {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, _ := m.BalanceOf(receiver) // No error expected.

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	m.holdings[id] = new(big.Int).Set(senderBal)
	m.holdings[receiver] = new(big.Int).Set(receiverBal)
	return nil
}

func (m MemAsset) BalanceOf(id string) (*big.Int, error) {
	current, ok := m.holdings[id]
	if !ok {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(current), nil
}

func NewMemAsset() *MemAsset {
	return &MemAsset{
		holdings: make(map[string]*big.Int),
	}
}
