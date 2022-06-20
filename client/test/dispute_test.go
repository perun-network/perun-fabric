//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package client_test

import (
	"context"
	"flag"
	"fmt"
	"github.com/perun-network/perun-fabric/channel"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	"github.com/perun-network/perun-fabric/wallet"
	"github.com/stretchr/testify/assert"
	"math/big"
	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	"testing"
	"time"
)

const (
	disputeTestTimeout = 60 * time.Second
	malloryHolding     = 1000
	carolHolding       = 1000
)

func TestDisputeMalloryCarol(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), disputeTestTimeout)
	defer cancel()

	flag.Parse()

	const (
		A, B = 0, 1 // Indices of Mallory and Carol
	)

	var (
		name  = [2]string{"Mallory", "Carol"}
		role  [2]clienttest.Executer
		setup [2]clienttest.RoleSetup
	)

	var adjs []*chtest.Session
	for i := uint(1); i <= 2; i++ {
		as, err := chtest.NewTestSession(chtest.OrgNum(i), chaincode)
		chtest.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	bus := wire.NewLocalBus()

	for i := 0; i < len(setup); i++ {
		watcher, _ := local.NewWatcher(adjs[i].Adjudicator)

		setup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          adjs[i].Account,
			Bus:               bus,
			Funder:            adjs[i].Funder,
			Adjudicator:       adjs[i].Adjudicator,
			Wallet:            wallet.NewWallet(),
			Timeout:           10 * time.Second, // Timeout waiting for other role, not challenge duration
			ChallengeDuration: 15,
			Watcher:           watcher,
		}
	}

	role[A] = clienttest.NewMallory(t, setup[A])
	role[B] = clienttest.NewCarol(t, setup[B])

	execConfig := &clienttest.MalloryCarolExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			channel.Asset,
			[2]*big.Int{big.NewInt(malloryHolding), big.NewInt(carolHolding)},
			pclient.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(50), big.NewInt(0)},
	}

	err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
	assert.NoError(t, err)
}
