// SPDX-License-Identifier: Apache-2.0

package adjudicator

// Adjudicator is an abstract implementation of the adjudicator smart
// contract.
type Adjudicator struct {
	ledger Ledger
	assets *AssetHolder
}

func NewAdjudicator(ledger Ledger) *Adjudicator {
	return &Adjudicator{
		ledger: ledger,
		assets: NewAssetHolder(ledger),
	}
}
