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

package adjudicator_test

import (
	"encoding/json"
	"github.com/perun-network/perun-fabric/wallet"
	"perun.network/go-perun/channel"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"
	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	adjtest "github.com/perun-network/perun-fabric/adjudicator/test"
)

func TestStateRegJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	sr := adjtest.RandomStateReg(rng)
	data, err := json.Marshal(sr)
	require.NoError(t, err)
	sr1 := new(adj.StateReg)
	require.NoError(t, json.Unmarshal(data, sr1))
	require.Zero(t, deep.Equal(sr, sr1))
}

func TestStateSigning(t *testing.T) {
	rng := test.Prng(t)
	state := adjtest.RandomState(rng)
	acc := wtest.NewRandomAccount(rng)
	sig, err := state.Sign(acc)
	require.NoError(t, err)
	ok, err := adj.VerifySig(acc.Address(), *state, sig)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSignedWithdrawRequestJSONMarshaling(t *testing.T) {
	rng := test.Prng(t)
	acc := wallet.NewRandomAccount(rng)
	req, err := adj.SignWithdrawRequest(acc, channel.ID{1}, "someReceiverID")
	require.NoError(t, err)
	data, err := json.Marshal(req)
	require.NoError(t, err)
	req1 := new(adj.SignedWithdrawReq)
	require.NoError(t, json.Unmarshal(data, req1))
	require.Zero(t, deep.Equal(req, req1))
}

func TestWithdrawRequestSigning(t *testing.T) {
	rng := test.Prng(t)
	acc := wallet.NewRandomAccount(rng)
	req, err := adj.SignWithdrawRequest(acc, channel.ID{1}, "someReceiverID")
	require.NoError(t, err)
	verify, err := req.Verify(acc.Address())
	require.NoError(t, err)
	require.True(t, verify)
}

func TestWithdrawRequestSigningInvalid(t *testing.T) {
	rng := test.Prng(t)
	acc := wallet.NewRandomAccount(rng)
	req, err := adj.SignWithdrawRequest(acc, channel.ID{1}, "someReceiverID")
	require.NoError(t, err)
	acc1 := wallet.NewRandomAccount(rng)
	verify, err := req.Verify(acc1.Address())
	require.NoError(t, err)
	require.False(t, verify)
}
