package helmutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHelm3(t *testing.T) {
	testCases := map[string]struct {
		setup   func() func()
		isHelm3 bool
	}{
		"TILLER_HOST is set": {
			setup: func() func() {
				return mockEnv(t, "TILLER_HOST", "1")
			},
			isHelm3: false,
		},
		"helm command detects v2": {
			setup: func() func() {
				helm3Detected = func() bool {
					return false
				}
				return mockEnv(t, "TILLER_HOST", "")
			},
			isHelm3: false,
		},
		"helm 3 version format": {
			setup: func() func() {
				helm3Detected = func() bool {
					return true
				}
				return mockEnv(t, "TILLER_HOST", "")
			},
			isHelm3: true,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			teardown := tc.setup()
			defer teardown()

			assert.Equal(t, tc.isHelm3, IsHelm3())
		})
	}
}
