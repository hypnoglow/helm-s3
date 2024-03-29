package helmutil

// This file contains helm helpers suitable for both v2 and v3.

import (
	"fmt"
	"strings"
)

func SetupHelm() {
	setupHelmVersionDetection()
	setupHelm2()
	setupHelm3()
}

// IndexFileURL returns index file URL for the provided repository URL.
func IndexFileURL(repoURL string) string {
	return strings.TrimSuffix(repoURL, "/") + "/index.yaml"
}

func repoCacheFileName(name string) string {
	return fmt.Sprintf("%s-index.yaml", name)
}
