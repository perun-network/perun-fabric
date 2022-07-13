package adjudicator

import (
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type Asset interface {
	Mint(identity string, addr wallet.Address, amount *big.Int) error
	Burn(identity string, addr wallet.Address, amount *big.Int) error

	AddressToAddressTransfer(identity string, sender wallet.Address, receiver wallet.Address, amount *big.Int) error
	AddressToChannelTransfer(identity string, sender wallet.Address, receiver channel.ID, amount *big.Int) error
	ChannelToAddressTransfer(sender channel.ID, receiver wallet.Address, amount *big.Int) error

	BalanceOfAddress(address wallet.Address) (*big.Int, error)
	BalanceOfChannel(id channel.ID) (*big.Int, error)

	RegisterAddress(identity string, addr wallet.Address) error
	GetAddressIdentity(addr wallet.Address) (string, error)
}
