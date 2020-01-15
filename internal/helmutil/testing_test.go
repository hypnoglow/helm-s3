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

func mockEnvs(t *testing.T, nameValue ...string) func() {
	if len(nameValue)%2 != 0 {
		t.Fatal("mockEnvs: must have even number of arguments")
	}

	tearDowns := make([]func(), 0, len(nameValue)/2)
	for i := 0; i < len(nameValue); i++ {
		if i%2 == 1 {
			continue
		}

		name := nameValue[i]
		value := nameValue[i+1]

		old := os.Getenv(name)

		err := os.Setenv(name, value)
		require.NoError(t, err)

		tearDowns = append(tearDowns, func() {
			err := os.Setenv(name, old)
			require.NoError(t, err)
		})
	}

	return func() {
		for _, td := range tearDowns {
			td()
		}
	}
}

func assertError(t *testing.T, err error, expected bool) {
	if expected {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}
