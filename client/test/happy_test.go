// Copyright 2021 PolyCrypt GmbH
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

package client_test

import (
	"context"
	"flag"
	"fmt"
	chtest "github.com/perun-network/perun-fabric/channel/test"
	wallet "github.com/perun-network/perun-fabric/wallet"
	"math/big"
	"perun.network/go-perun/watcher/local"
	"testing"
	"time"

	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"

	"github.com/perun-network/perun-fabric/channel"
	"github.com/stretchr/testify/assert"
)

const (
	testTimeout  = 60 * time.Second
	aliceHolding = 100000000
	bobHolding   = 100000000
)

var adjudicator = flag.String("adjudicator", "adjudicator-23465", "Adjudicator chaincode name")
var assetholder = flag.String("assetholder", "assetholder-23465", "AssetHolder chaincode name")

func TestHappyAliceBob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	flag.Parse()

	const (
		A, B = 0, 1 // Indices of Alice and Bob
	)
	var (
		name  = [2]string{"Alice", "Bob"}
		role  [2]clienttest.Executer
		setup [2]clienttest.RoleSetup
	)

	var adjs []*chtest.Session
	for i := uint(1); i <= 2; i++ {
		as, err := chtest.NewTestSession(chtest.OrgNum(i), *adjudicator, *assetholder)
		chtest.FatalErr(fmt.Sprintf("creating adjudicator session[%d]", i), err)
		defer as.Close()
		adjs = append(adjs, as)
	}

	bus := wire.NewLocalBus()

	for i := 0; i < len(setup); i++ {
		acc := adjs[i].Account
		//wallet := wallet.NewWallet(acc)
		watcher, _ := local.NewWatcher(adjs[i].Adjudicator)

		setup[i] = clienttest.RoleSetup{
			Name:              name[i],
			Identity:          acc,
			Bus:               bus,
			Funder:            adjs[i].Funder,
			Adjudicator:       adjs[i].Adjudicator,
			Wallet:            wallet.NewWallet(),
			Timeout:           60 * time.Second, // Timeout waiting for other role, not challenge duration
			ChallengeDuration: 60,
			Watcher:           watcher,
		}
	}

	//setup := makeRoleSetups(name, *adjudicator, *assetholder)
	role[A] = clienttest.NewAlice(t, setup[A])
	role[B] = clienttest.NewBob(t, setup[B])

	fmt.Printf("Client 0: %s, Client 1: %s \n", setup[A].Identity.Address().String(), setup[B].Identity.Address().String())

	execConfig := &clienttest.AliceBobExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			channel.Asset,
			[2]*big.Int{big.NewInt(aliceHolding), big.NewInt(bobHolding)},
			pclient.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(20), big.NewInt(0)},
	}

	// Amount that will be send from Alice to Bob.
	//aliceToBob := big.NewInt(int64(execConfig.NumPayments[A])*execConfig.TxAmounts[A].Int64() - int64(execConfig.NumPayments[B])*execConfig.TxAmounts[B].Int64())
	// Amount that will be send from Bob to Alice.
	//bobToAlice := new(big.Int).Neg(aliceToBob)
	// Expected balance changes of the accounts.
	//deltas := map[types.AccountID]*big.Int{
	//	wallet.AsAddr(s.Alice.Acc.Address()).AccountID(): aliceToBob,
	//	wallet.AsAddr(s.Bob.Acc.Address()).AccountID():   bobToAlice,
	//}
	//s.AssertBalanceChanges(deltas, epsilon, func() {
	err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
	assert.NoError(t, err)
	//})
}
