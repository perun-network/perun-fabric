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
		timeout := time.Now().UTC().Add(duration)
		t0 := channel.MakeTimeout(timeout, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		time.Sleep(duration)
		assert.True(t, t0.IsElapsed(context.Background()))
	})

	t.Run("Wait", func(t *testing.T) {
		timeout := time.Now().UTC().Add(duration)
		t0 := channel.MakeTimeout(timeout, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		err := t0.Wait(context.Background())
		assert.NoError(t, err)
		assert.True(t, t0.IsElapsed(context.Background()))
	})

	t.Run("Wait Error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), duration/2)
		defer cancel()

		timeout := time.Now().UTC().Add(duration)
		t0 := channel.MakeTimeout(timeout, polling)
		assert.False(t, t0.IsElapsed(context.Background()))

		err := t0.Wait(ctx)
		assert.Error(t, ctx.Err(), err)
	})
}
