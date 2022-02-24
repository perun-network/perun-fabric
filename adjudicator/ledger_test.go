package adjudicator_test

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"
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
