// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"testing"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/stretchr/testify/assert"
)

func TestStdTimestamp(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		t0 := adj.StdNow()
		t1 := t0.Add(1)
		assert.True(t, t1.After(t0))
	})

	t0 := adj.StdNow()

	t.Run("Clone", func(t *testing.T) {
		t0c := t0.Clone()
		t0c = t0c.Add(1)
		assert.False(t, t0c.Equal(t0))
	})

	t1 := adj.StdNow()

	t.Run("Equal", func(t *testing.T) {
		assert.True(t, t0.Equal(t0))
		assert.True(t, t1.Equal(t1))
		assert.False(t, t0.Equal(t1))
	})

	t.Run("After", func(t *testing.T) {
		assert.False(t, t0.After(t1))
		assert.False(t, t0.After(t0))
		assert.True(t, t1.After(t0))
	})

	t.Run("Before", func(t *testing.T) {
		assert.True(t, t0.Before(t1))
		assert.False(t, t0.Before(t0))
		assert.False(t, t1.Before(t0))
	})

	t.Run("Marshalling", func(t *testing.T) {
		mar, err := t0.MarshalJSON()
		assert.NoError(t, err)
		t2 := adj.StdNow()
		assert.True(t, t0.Before(t2))
		err = t2.UnmarshalJSON(mar)
		assert.NoError(t, err)
		assert.True(t, t0.Equal(t2))
	})
}
