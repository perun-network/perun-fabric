// SPDX-License-Identifier: Apache-2.0

package wallet

import (
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
)

func init() {
	wallet.SetBackend(Backend{})
	wtest.SetRandomizer(NewRandomizer())
}
