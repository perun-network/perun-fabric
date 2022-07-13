package chaincode

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type StubAsset struct {
	Stub shim.ChaincodeStubInterface
}

func NewStubAsset(ctx contractapi.TransactionContextInterface) *StubAsset {
	return &StubAsset{Stub: ctx.GetStub()}
}

func (s StubAsset) Mint(identity string, addr wallet.Address, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check if identity registered addr.
	err := s.checkIdentity(identity, addr)
	if err != nil {
		return err
	}
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOfAddress(addr)
	if err != nil {
		return err
	}
	current.Add(current, amount)
	if err := s.Stub.PutState(addressKey(addr), current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) Burn(identity string, addr wallet.Address, amount *big.Int) error {
	// Check if callee == centralBanker. Skipped for this demo asset.
	// Check if identity registered addr.
	err := s.checkIdentity(identity, addr)
	if err != nil {
		return err
	}
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot burn zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOfAddress(addr)
	if err != nil {
		return err
	}
	current.Sub(current, amount)

	if current.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("not enought funds to burn the requested amount")
	}

	if err := s.Stub.PutState(addressKey(addr), current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) AddressToAddressTransfer(identity string, sender wallet.Address, receiver wallet.Address, amount *big.Int) error {
	// Check if identity registered sender address.
	err := s.checkIdentity(identity, sender)
	if err != nil {
		return err
	}
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, err := s.BalanceOfAddress(sender)
	if err != nil {
		return err
	}
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, err := s.BalanceOfAddress(receiver)
	if err != nil {
		return err
	}

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	if err := s.Stub.PutState(addressKey(sender), senderBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	if err := s.Stub.PutState(addressKey(receiver), receiverBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) AddressToChannelTransfer(identity string, sender wallet.Address, receiver channel.ID, amount *big.Int) error {
	// Check if identity registered sender address.
	err := s.checkIdentity(identity, sender)
	if err != nil {
		return err
	}
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, err := s.BalanceOfAddress(sender)
	if err != nil {
		return err
	}
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, err := s.BalanceOfChannel(receiver)
	if err != nil {
		return err
	}

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	if err := s.Stub.PutState(addressKey(sender), senderBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	if err := s.Stub.PutState(channelKey(receiver), receiverBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) ChannelToAddressTransfer(sender channel.ID, receiver wallet.Address, amount *big.Int) error {
	// Validate that receiver address is registered.
	receiverID, err := s.GetAddressIdentity(receiver)
	if err != nil {
		return err
	} else if receiverID == "" {
		return fmt.Errorf("receiver address %s not registered", receiver.String())
	}
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot transfer zero/negative amount")
	}

	// Get balances of parties.
	senderBal, err := s.BalanceOfChannel(sender)
	if err != nil {
		return err
	}
	if senderBal.Cmp(amount) < 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, err := s.BalanceOfAddress(receiver)
	if err != nil {
		return err
	}

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	if err := s.Stub.PutState(channelKey(sender), senderBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	if err := s.Stub.PutState(addressKey(receiver), receiverBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) BalanceOfAddress(address wallet.Address) (*big.Int, error) {
	srb, err := s.Stub.GetState(addressKey(address))
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).SetBytes(srb), nil
}

func (s StubAsset) BalanceOfChannel(id channel.ID) (*big.Int, error) {
	srb, err := s.Stub.GetState(channelKey(id))
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).SetBytes(srb), nil
}

func (s StubAsset) RegisterAddress(identity string, addr wallet.Address) error {
	owner, err := s.GetAddressIdentity(addr)
	if err != nil {
		return err
	}

	// Skip double registrations.
	if owner == identity {
		return nil
	} else if owner != "" && owner != identity { // Already registered.
		return fmt.Errorf("%s is already registered", addr.String())
	}
	if err := s.Stub.PutState(identityKey(addr), []byte(identity)); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

// GetAddressIdentity checks if an identity registered the given address and return it.
// If address is unregistered an empty string is returned.
func (s StubAsset) GetAddressIdentity(addr wallet.Address) (string, error) {
	srb, err := s.Stub.GetState(identityKey(addr))
	if err != nil {
		return "", fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return "", nil
	}
	owner := bytes.NewBuffer(srb).String()
	return owner, nil
}

// checkIdentity validates that the given identity registered
// the given address and is therefore allowed to use it as its owner.
func (s StubAsset) checkIdentity(identity string, addr wallet.Address) error {
	addrId, err := s.GetAddressIdentity(addr)
	if err != nil {
		return err
	} else if addrId != identity {
		return fmt.Errorf("identity %s did not register address %s - identity %s is owner", identity, addr.String(), addrId)
	}
	return nil
}

func identityKey(address wallet.Address) string {
	return orgPrefix + "Identity:" + address.String()
}

// assetKey (id, addr) mapping:
// 1. User assets: ("", user addr)
// 2. Channel assets: (channel id, "")
func assetKey(id string, addr string) string {
	return orgPrefix + "AssetHolding:" + fmt.Sprintf("%s:%s", id, addr)
}

func addressKey(addr wallet.Address) string {
	return assetKey("", addr.String())
}

func channelKey(id channel.ID) string {
	idStr := hex.EncodeToString(id[:])
	return assetKey(idStr, "")
}
