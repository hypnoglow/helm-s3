package helmutil

// this file contains utilities for testing code in this package.

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockEnv(t *testing.T, name, value string) func() {
	old := os.Getenv(name)

	err := os.Setenv(name, value)
	require.NoError(t, err)

	return func() {
		err := os.Setenv(name, old)
		require.NoError(t, err)
	}
}

func assertError(t *testing.T, err error, expected bool) {
	if expected {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}
