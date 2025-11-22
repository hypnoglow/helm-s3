package helmutil

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "helm", "version", "--short", "--client")
	out, err := cmd.Output()
	if err != nil {
		// Should not happen in normal cases (when helm is properly installed).
		// Anyway, fallback to v3 since helm v2 was deprecated a long time ago
		// and the majority of helm-s3 users use v3.
		return true
	}

	return strings.HasPrefix(string(out), "v3.")
}

// setupHelmVersionDetection sets up the command used to detect helm version.
func setupHelmVersionDetection() {
	helm3Detected = helmVersionCommand
}
