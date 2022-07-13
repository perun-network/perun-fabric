package adjudicator

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type MemAsset struct {
	holdings map[string]*big.Int
}

func (m MemAsset) Mint(identity string, addr wallet.Address, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	current, _ := m.BalanceOfAddress(addr) // No error expected
	current.Add(current, amount)
	m.holdings[AddressKey(addr)] = new(big.Int).Set(current)
	return nil
}

func (m MemAsset) Burn(identity string, addr wallet.Address, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot burn zero/negative amount")
	}

	// Get current balance.
	current, _ := m.BalanceOfAddress(addr) // No error expected.
	current.Sub(current, amount)

	if current.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("not enought funds to burn the requested amount")
	}

	m.holdings[AddressKey(addr)] = new(big.Int).Set(current)
	return nil
}

func (m MemAsset) AddressToAddressTransfer(identity string, sender wallet.Address, receiver wallet.Address, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, _ := m.BalanceOfAddress(sender) // No error expected.
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, _ := m.BalanceOfAddress(receiver) // No error expected.

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	m.holdings[AddressKey(sender)] = new(big.Int).Set(senderBal)
	m.holdings[AddressKey(receiver)] = new(big.Int).Set(receiverBal)
	return nil
}

func (m MemAsset) AddressToChannelTransfer(identity string, sender wallet.Address, receiver channel.ID, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, _ := m.BalanceOfAddress(sender) // No error expected.
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, _ := m.BalanceOfChannel(receiver) // No error expected.

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	m.holdings[AddressKey(sender)] = new(big.Int).Set(senderBal)
	m.holdings[ChannelKey(receiver)] = new(big.Int).Set(receiverBal)
	return nil
}

func (m MemAsset) ChannelToAddressTransfer(sender channel.ID, receiver wallet.Address, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, _ := m.BalanceOfChannel(sender) // No error expected.
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, _ := m.BalanceOfAddress(receiver) // No error expected.

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	m.holdings[ChannelKey(sender)] = new(big.Int).Set(senderBal)
	m.holdings[AddressKey(receiver)] = new(big.Int).Set(receiverBal)
	return nil
}

func (m MemAsset) BalanceOfAddress(address wallet.Address) (*big.Int, error) {
	current, ok := m.holdings[AddressKey(address)]
	if !ok {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(current), nil
}

func (m MemAsset) BalanceOfChannel(id channel.ID) (*big.Int, error) {
	current, ok := m.holdings[ChannelKey(id)]
	if !ok {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(current), nil
}

// RegisterAddress skipped for MemAsset.
func (m MemAsset) RegisterAddress(identity string, addr wallet.Address) error {
	return nil
}

// GetAddressIdentity skipped for MemAsset.
func (m MemAsset) GetAddressIdentity(addr wallet.Address) (string, error) {
	return "", nil
}

// assetKey (id, addr) mapping:
// 1. User assets: ("", user addr)
// 2. Channel assets: (channel id, "")
func assetKey(id string, addr string) string {
	return fmt.Sprintf("%s:%s", id, addr)
}

func AddressKey(addr wallet.Address) string {
	return assetKey("", addr.String())
}

func ChannelKey(id channel.ID) string {
	idStr := hex.EncodeToString(id[:])
	return assetKey(idStr, "")
}

func NewMemAsset() *MemAsset {
	return &MemAsset{
		holdings: make(map[string]*big.Int),
	}
}
