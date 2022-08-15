// SPDX-License-Identifier: Apache-2.0

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
