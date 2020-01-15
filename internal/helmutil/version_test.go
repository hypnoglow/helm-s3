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
		"HELM_S3_MODE is set to 2": {
			setup: func() func() {
				return mockEnv(t, "HELM_S3_MODE", "2")
			},
			isHelm3: false,
		},
		"HELM_S3_MODE is set to 3": {
			setup: func() func() {
				return mockEnv(t, "HELM_S3_MODE", "v3")
			},
			isHelm3: true,
		},
		"HELM_S3_MODE is set to any other value, TILLER_HOST is empty, helm command detects v2": {
			setup: func() func() {
				helm3Detected = func() bool {
					return false
				}
				return mockEnvs(t,
					"HELM_S3_MODE", "abc",
					"TILLER_HOST", "",
				)
			},
			isHelm3: false,
		},
		"HELM_S3_MODE is empty, TILLER_HOST is empty, helm command detects v2": {
			setup: func() func() {
				helm3Detected = func() bool {
					return false
				}
				return mockEnvs(t,
					"HELM_S3_MODE", "",
					"TILLER_HOST", "",
				)
			},
			isHelm3: false,
		},
		"HELM_S3_MODE is empty, TILLER_HOST is empty, helm command detects v3": {
			setup: func() func() {
				helm3Detected = func() bool {
					return true
				}
				return mockEnvs(t,
					"HELM_S3_MODE", "",
					"TILLER_HOST", "",
				)
			},
			isHelm3: true,
		},
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
