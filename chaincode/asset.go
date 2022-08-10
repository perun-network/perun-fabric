package chaincode

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"math/big"
)

// centralBankerID is the user allowed to mint & burn tokens.
const centralBankerID = "eDUwOTo6Q049dXNlcjEsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw=="

type StubAsset struct {
	Stub shim.ChaincodeStubInterface
}

func NewStubAsset(ctx contractapi.TransactionContextInterface) *StubAsset {
	return &StubAsset{Stub: ctx.GetStub()}
}

func (s StubAsset) Mint(id string, amount *big.Int) error {
	// Check if id == centralBanker.
	if id != centralBankerID {
		return fmt.Errorf("minter must be central banker")
	}

	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot mint zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOf(id)
	if err != nil {
		return err
	}
	current.Add(current, amount)
	if err := s.Stub.PutState(id, current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) Burn(id string, amount *big.Int) error {
	// Check if id == centralBanker.
	if id != centralBankerID {
		return fmt.Errorf("burner must be central banker")
	}

	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("cannot burn zero/negative amount")
	}

	// Get current balance.
	current, err := s.BalanceOf(id)
	if err != nil {
		return err
	}
	current.Sub(current, amount)

	if current.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("not enought funds to burn the requested amount")
	}

	if err := s.Stub.PutState(id, current.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) Transfer(id string, receiver string, amount *big.Int) error {
	// Check zero/negative amount.
	if amount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("cannot transfer negative amount")
	}

	// Get balances of parties.
	senderBal, err := s.BalanceOf(id)
	if err != nil {
		return err
	}
	if senderBal.Cmp(amount) <= 0 {
		return fmt.Errorf("not enought funds to transfer the requested amount")
	}
	receiverBal, err := s.BalanceOf(receiver)
	if err != nil {
		return err
	}

	// Calc new balances.
	senderBal.Sub(senderBal, amount)
	receiverBal.Add(receiverBal, amount)

	// Store new balances.
	if err := s.Stub.PutState(id, senderBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	if err := s.Stub.PutState(receiver, receiverBal.Bytes()); err != nil {
		return fmt.Errorf("stub.PutState: %w", err)
	}
	return nil
}

func (s StubAsset) BalanceOf(id string) (*big.Int, error) {
	srb, err := s.Stub.GetState(id)
	if err != nil {
		return nil, fmt.Errorf("stub.GetState: %w", err)
	} else if srb == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).SetBytes(srb), nil
}
