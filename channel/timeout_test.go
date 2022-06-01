package channel_test

import (
	"context"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	duration := 50 * time.Millisecond
	polling := duration

	t.Run("Elapsed", func(t *testing.T) {

		t0 := channel.MakeTimeout(duration, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		time.Sleep(duration)
		assert.True(t, t0.IsElapsed(context.Background()))
	})

	t.Run("Wait", func(t *testing.T) {
		t0 := channel.MakeTimeout(duration, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		err := t0.Wait(context.Background())
		assert.NoError(t, err)
		assert.True(t, t0.IsElapsed(context.Background()))
	})

	t.Run("Wait Error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), duration/2)
		defer cancel()

		t0 := channel.MakeTimeout(duration, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		err := t0.Wait(ctx)
		assert.Error(t, ctx.Err(), err)
	})
}
