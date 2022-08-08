// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"encoding/json"
	"fmt"

	"perun.network/go-perun/wallet"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	_ "github.com/perun-network/perun-fabric" // init backend
	fabwallet "github.com/perun-network/perun-fabric/wallet"
)

func stringWithErr(s fmt.Stringer, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

func UnmarshalID(idStr string) (string, error) {
	id := ""
	if err := json.Unmarshal([]byte(idStr), &id); err != nil {
		return id, fmt.Errorf("json-unmarshaling Client ID: %w", err)
	}
	return id, nil
}

func UnmarshalAddress(addrStr string) (wallet.Address, error) {
	addr := wallet.NewAddress()
	if err := json.Unmarshal([]byte(addrStr), &addr); err != nil {
		return addr, fmt.Errorf("json-unmarshaling Farbic Address: %w", err)
	}
	return addr, nil
}

func UnmarshalAddresses(addrsStr string) ([]wallet.Address, error) {
	var jsonAddrs []json.RawMessage
	if err := json.Unmarshal([]byte(addrsStr), &jsonAddrs); err != nil {
		return nil, fmt.Errorf("unmarshaling array: %w", err)
	}

	addrs := make([]wallet.Address, 0, len(jsonAddrs))
	for i, addrStr := range jsonAddrs {
		addr, err := UnmarshalAddress(string(addrStr))
		if err != nil {
			return nil, fmt.Errorf("unmarshal[%d]: %w", i, err)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func AddressFromTxCtx(ctx contractapi.TransactionContextInterface) (*fabwallet.Address, error) {
	cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return nil, fmt.Errorf("getting identity: %w", err)
	}
	return fabwallet.AddressFromX509Certificate(cert)
}
