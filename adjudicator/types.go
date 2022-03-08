package adjudicator

import (
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"

	_ "perun.network/go-perun/backend/ethereum"
)

type (
	Params struct {
		ChallengeDuration uint64        `json:"challengeDuration"`
		Parts             []Address     `json:"parts"`
		Nonce             channel.Nonce `json:"nonce"`
	}

	State struct {
		ID       channel.ID    `json:"id"`
		Version  uint64        `json:"version"`
		Balances []channel.Bal `json:"balances"`
		IsFinal  bool          `json:"final"`
	}

	Address = wallet.Address
)
