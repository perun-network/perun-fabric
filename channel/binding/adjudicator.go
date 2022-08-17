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
	txTToAddr           = "TransferToken"
	txTBal              = "TokenBalance"
	submitRetryDuration = 3 * time.Second
)

// Adjudicator wraps a fabric client.Contract to connect to the Adjudicator chaincode.
type Adjudicator struct {
	Contract *client.Contract
}

// NewAdjudicatorBinding creates the bindings for the on-chain Adjudicator.
// These bindings are the main point of interaction with the chaincode.
func NewAdjudicatorBinding(network *client.Network, chainCode string) *Adjudicator {
	return &Adjudicator{Contract: network.GetContract(chainCode)}
}

// Deposit marshals the given parameters and sends a deposits request to the Adjudicator chaincode.
func (a *Adjudicator) Deposit(id channel.ID, part wallet.Address, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, part, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txDeposit, args...)
	return err
}

// Holding marshals the given parameters and sends a holding request to the Adjudicator chaincode.
// The response contains the current holding of the given address in the channel.
func (a *Adjudicator) Holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txHolding, args...))
}

// TotalHolding marshals the given parameters and sends a total holding request to the Adjudicator chaincode.
// The response contains the sum of the current holdings of the given addresses in the channel.
func (a *Adjudicator) TotalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txTotalHolding, args...))
}

// Register marshals the signed channel state and sends a register request to the Adjudicator chaincode.
func (a *Adjudicator) Register(ch *adj.SignedChannel) error {
	arg, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txRegister, string(arg))
	return err
}

// StateReg marshals the given channel id and sends a state reg request to the Adjudicator chaincode.
// The response contains the current registered state of the given channel.
func (a *Adjudicator) StateReg(id channel.ID) (*adj.StateReg, error) {
	arg, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}
	regJSON, err := a.submitTransactionWithRetry(txStateReg, string(arg))
	if err != nil {
		return nil, err
	}
	var reg adj.StateReg
	return &reg, json.Unmarshal(regJSON, &reg)
}

// Withdraw marshals the given withdraw request and sends it to the Adjudicator chaincode.
// The response contains the amount of funds withdrawn form the channel.
func (a *Adjudicator) Withdraw(req adj.SignedWithdrawReq) (*big.Int, error) {
	arg, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txWithdraw, string(arg)))
}

// MintToken marshals the given amount and sends a request to the Adjudicator chaincode to mint the amount of tokens.
func (a *Adjudicator) MintToken(amount *big.Int) error {
	arg, err := json.Marshal(amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txMintT, string(arg))
	return err
}

// BurnToken marshals the given amount and sends a request to the Adjudicator chaincode to burn the amount of tokens.
func (a *Adjudicator) BurnToken(amount *big.Int) error {
	arg, err := json.Marshal(amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txBurnT, string(arg))
	return err
}

// TokenTransfer marshals the given parameters and sends a token transfer request to the Adjudicator chaincode.
func (a *Adjudicator) TokenTransfer(receiver adj.AccountID, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(receiver, amount)
	if err != nil {
		return err
	}
	_, err = a.submitTransactionWithRetry(txTToAddr, args...)
	return err
}

// TokenBalance marshals the given owner id and sends a token balance request to the Adjudicator chaincode.
// The response contains the amount of tokens the given owner id holds.
func (a *Adjudicator) TokenBalance(owner adj.AccountID) (*big.Int, error) {
	arg, err := json.Marshal(owner)
	if err != nil {
		return nil, err
	}
	return bigIntWithError(a.submitTransactionWithRetry(txTBal, string(arg)))
}

// submitTransactionWithRetry ensures that in case of a missed lock on the contract there is
// another attempt on submitting the transaction.
func (a *Adjudicator) submitTransactionWithRetry(txType string, args ...string) ([]byte, error) {
	tx, err := a.Contract.SubmitTransaction(txType, args...)
	if e, ok := err.(*client.CommitError); ok && e.Code == peer.TxValidationCode_MVCC_READ_CONFLICT { //nolint:nosnakecase
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
