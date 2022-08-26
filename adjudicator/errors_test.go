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
	"errors"
	"fmt"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsAdjudicatorError(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		require.False(t, adj.IsAdjudicatorError(nil))
		noerr := errors.New("no adjudicator error")
		require.False(t, adj.IsAdjudicatorError(noerr))
	})

	adjErrors := []error{
		adj.ValidationError{},
		adj.ChallengeTimeoutError{},
		adj.VersionError{},
		adj.UnderfundedError{},
	}

	t.Run("true", func(t *testing.T) {
		for _, err := range adjErrors {
			require.True(t, adj.IsAdjudicatorError(err))
		}
	})

	t.Run("wrapped", func(t *testing.T) {
		for _, err := range adjErrors {
			werr := fmt.Errorf("wrap: %w", err)
			require.True(t, adj.IsAdjudicatorError(werr))
		}
	})
}
