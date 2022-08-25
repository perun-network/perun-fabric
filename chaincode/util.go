// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
