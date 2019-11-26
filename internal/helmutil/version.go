package helmutil

import (
	"os"
	"os/exec"
)

// IsHelm3 returns true if helm is version 3+.
func IsHelm3() bool {
	if os.Getenv("TILLER_HOST") != "" {
		return false
	}

	return helm3Detected()
}

// helm3Detected returns true if helm is v3.
var helm3Detected func() bool

func helmEnvCommand() bool {
	cmd := exec.Command("helm", "env")
	return cmd.Run() == nil
}

// setupHelmVersionDetection sets up the command used to detect helm version.
func setupHelmVersionDetection() {
	helm3Detected = helmEnvCommand
}
