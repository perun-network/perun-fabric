// SPDX-License-Identifier: Apache-2.0

package adjudicator_test

import (
	"testing"
	"time"

	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/stretchr/testify/assert"
)

func TestStdTime(t *testing.T) {
	t0, t1 := adj.StdNow(), adj.StdNow()

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

	t.Run("Add", func(t *testing.T) {
		t0 := adj.StdNow()
		t1 := t0.Add(1)
		assert.True(t, t1.After(t0))
	})

	t.Run("Clone", func(t *testing.T) {
		t0c := t0.Clone().(*adj.StdTimestamp)
		(*t0c) = (adj.StdTimestamp)(t0c.Time().Add(time.Hour))
		assert.False(t, t0c.Equal(t0))
	})
}
