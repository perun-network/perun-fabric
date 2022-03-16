package adjudicator_test

import (
	"encoding/json"
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
