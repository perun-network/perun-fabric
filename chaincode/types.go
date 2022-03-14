// SPDX-License-Identifier: Apache-2.0

package chaincode

import (
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"

	adj "github.com/perun-network/perun-fabric/adjudicator"
)

type (
	// Address is the concrete address type used.
	// Unfortunately, this has to be set because fabric-gateway has strong
	// limitations on what types can cross the chaincode API boundary.
	Address = ethwallet.Address

	// RegTimestamp is the concrete timestamp type used in the registry.
	// Unfortunately, this has to be set because fabric-gateway has strong
	// limitations on what types can cross the chaincode API boundary.
	RegTimestamp = adj.StdTimestamp
)

func AsWalletAddresses(as []Address) []wallet.Address {
	was := make([]wallet.Address, 0, len(as))
	for _, a := range as {
		was = append(was, &a)
	}
	return was
}
