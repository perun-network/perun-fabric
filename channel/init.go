// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"perun.network/go-perun/channel"

	chtest "perun.network/go-perun/channel/test"
)

func init() {
	channel.SetBackend(Backend{})
	chtest.SetRandomizer(Backend{})
}
