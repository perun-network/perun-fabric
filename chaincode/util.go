// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	"encoding/json"
	"fmt"

	"perun.network/go-perun/wallet"

	_ "github.com/perun-network/perun-fabric" // init backend
)

func stringWithErr(s fmt.Stringer, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

func UnmarshalAddress(addrStr string) (wallet.Address, error) {
	addr := wallet.NewAddress()
	if err := json.Unmarshal([]byte(addrStr), addr); err != nil {
		return nil, fmt.Errorf("json-unmarshaling Address: %w", err)
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
