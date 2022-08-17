// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"encoding/json"
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"

	"perun.network/go-perun/wallet"

	_ "github.com/perun-network/perun-fabric" // init backend
)

func stringWithErr(s fmt.Stringer, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

// UnmarshalID unmarshalls a fabric ID.
func UnmarshalID(idStr string) (adj.AccountID, error) {
	id := ""
	if err := json.Unmarshal([]byte(idStr), &id); err != nil {
		return adj.AccountID(id), fmt.Errorf("json-unmarshaling Client ID: %w", err)
	}
	return adj.AccountID(id), nil
}

// UnmarshalAddress implements custom unmarshalling of wallet addresses.
func UnmarshalAddress(addrStr string) (wallet.Address, error) {
	addr := wallet.NewAddress()
	if err := json.Unmarshal([]byte(addrStr), &addr); err != nil {
		return addr, fmt.Errorf("json-unmarshaling Farbic Address: %w", err)
	}
	return addr, nil
}

// UnmarshalAddresses unmarshalls an array of wallet addresses.
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
