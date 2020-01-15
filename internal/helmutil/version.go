package helmutil

import (
	"os"
	"os/exec"
	"strings"
)

// IsHelm3 returns true if helm is version 3+.
func IsHelm3() bool {
	// Support explicit mode configuration via environment variable.
	switch strings.TrimSpace(os.Getenv("HELM_S3_MODE")) {
	case "2", "v2":
		return false
	case "3", "v3":
		return true
	default:
		// continue to other detection methods.
	}

	if os.Getenv("TILLER_HOST") != "" {
		return false
	}

	return helm3Detected()
}

// helm3Detected returns true if helm is v3.
var helm3Detected func() bool

func helmVersionCommand() bool {
	cmd := exec.Command("helm", "version", "--short", "--client")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Should not happen in normal cases (when helm is properly installed).
		// Anyway, for now fallback to v2 for backward compatibility for helm-s3 users that are still on v2.
		return false
	}

	return strings.HasPrefix(string(out), "v3.")
}

// setupHelmVersionDetection sets up the command used to detect helm version.
func setupHelmVersionDetection() {
	helm3Detected = helmVersionCommand
}
