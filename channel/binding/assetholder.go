package binding

import (
	"math/big"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

	pkgjson "github.com/perun-network/perun-fabric/pkg/json"
	"github.com/perun-network/perun-fabric/tests"
)

type AssetHolder struct {
	Contract *client.Contract
}

func NewAssetHolderBinding(network *client.Network, chainCode string) *AssetHolder {
	return &AssetHolder{Contract: network.GetContract(chainCode)}
}

func (ah *AssetHolder) Deposit(id channel.ID, amount *big.Int) error {
	args, err := pkgjson.MultiMarshal(id, amount)
	if err != nil {
		return err
	}
	_, err = ah.Contract.SubmitTransaction(txDeposit, args...)
	return err
}

func (ah *AssetHolder) Holding(id channel.ID, addr wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addr)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(ah.Contract.SubmitTransaction(txHolding, args...))
}

func (ah *AssetHolder) TotalHolding(id channel.ID, addrs []wallet.Address) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id, addrs)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(ah.Contract.SubmitTransaction(txTotalHolding, args...))
}

func (ah *AssetHolder) Withdraw(id channel.ID) (*big.Int, error) {
	args, err := pkgjson.MultiMarshal(id)
	if err != nil {
		return nil, err
	}
	return tests.BigIntWithError(ah.Contract.SubmitTransaction(txWithdraw, args...))
}
