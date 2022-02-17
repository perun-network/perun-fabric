// SPDX-License-Identifier: Apache-2.0

package adjudicator

// Adjudicator is an abstract implementation of the adjudicator smart
// contract.
type Adjudicator struct {
	ledger Ledger
	assets *AssetHolder
}
