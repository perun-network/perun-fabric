package binding

import (
	"encoding/json"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go/peer"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	pkgjson "github.com/perun-network/perun-fabric/pkg/json"
	"math/big"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"time"
)

const (
	txDeposit           = "Deposit"
	txHolding           = "Holding"
	txTotalHolding      = "TotalHolding"
	txRegister          = "Register"
	txStateReg          = "StateReg"
	txWithdraw          = "Withdraw"
	txMintT             = "MintToken"
	txBurnT             = "BurnToken"
	txTToAddr           = "TokenToAddressTransfer"
	txTBal              = "TokenBalance"
	txRegAddr           = "RegisterAddress"
	txGetId             = "GetAddressIdentity"
	submitRetryDuration = 1 * time.Second
)

type Adjudicator struct {
	Contract *client.Contract
}

// NewAdjudicatorBinding creates the bindings for the on-chain Adjudicator.
// These bindings are the main point of interaction with the chaincode.
func NewAdjudicatorBinding(network *client.Network, chainCode string) *Adjudicator {
	return &Adjudicator{Contract: network.GetContract(chainCode)}
}

func (a *Adjudicator) Deposit(id channel.ID, part wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, part, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txDeposit, args...)
	return err
}

func (a *Adjudicator) Holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txHolding, args...))
}

func (a *Adjudicator) TotalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txTotalHolding, args...))
}

func (a *Adjudicator) Register(ch *adj.SignedChannel) error {
	arg, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txRegister, string(arg))
	return err
}

func (a *Adjudicator) StateReg(id channel.ID) (*adj.StateReg, error) {
	arg, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	regJson, err := a.submitTransactionWithRetry(txStateReg, string(arg))
	if err != nil {
		return nil, err
	}
	var reg adj.StateReg
	return &reg, json.Unmarshal(regJson, &reg)
}

func (a *Adjudicator) Withdraw(id channel.ID, part wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, part)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txWithdraw, args...))
}

func (a *Adjudicator) MintToken(addr wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(addr, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txMintT, args...)
	return err
}

func (a *Adjudicator) BurnToken(addr wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(addr, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txBurnT, args...)
	return err
}

func (a *Adjudicator) TokenToAddressTransfer(sender wallet.Address, receiver wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(sender, receiver, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txTToAddr, args...)
	return err
}

func (a *Adjudicator) TokenBalance(addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(addr)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txTBal, args...))
}

func (a *Adjudicator) RegisterAddress(addr wallet.Address) error {
	args, err := pkgjson.MultiMarshal(addr)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txRegAddr, args...)
	return err
}

func (a *Adjudicator) GetAddressIdentity(addr wallet.Address) (string, error) {
	args, err := pkgjson.MultiMarshal(addr)
	if err != nil {
		return "", err
	}
	return stringWithError(a.submitTransactionWithRetry(txGetId, args...))
}

// submitTransactionWithRetry ensures that in case of a missed lock on the contract there is
// another attempt on submitting the transaction
func (a *Adjudicator) submitTransactionWithRetry(txType string, args ...string) ([]byte, error) {
	tx, err := a.Contract.SubmitTransaction(txType, args...)
	if e, ok := err.(*client.CommitError); ok && e.Code == peer.TxValidationCode_MVCC_READ_CONFLICT {
		time.Sleep(submitRetryDuration)
		tx, err = a.Contract.SubmitTransaction(txType, args...)
	}
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func bigIntWithError(b []byte, err error) (*big.Int, error) {
	if err != nil {
		return nil, err
	}

	bi := new(big.Int)
	return bi, json.Unmarshal(b, bi)
}

func stringWithError(b []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}

	str := new(string)
	return *str, json.Unmarshal(b, str)
}
