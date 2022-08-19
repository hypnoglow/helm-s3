package main

import (
	"os"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

var (
	version = "master"
)

func main() {
	helmutil.SetupHelm()

	cmd := newRootCmd()

	if detectDownload(cmd, os.Args) {
		return
	}

	if err := cmd.Execute(); err != nil {
		if errorTypeSilent.Is(err) {
			os.Exit(1)
		}

		cmd.PrintErrln("Error:", err.Error())

		if errorTypeBadUsage.Is(err) {
			cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())
		}

		os.Exit(1)
	}
}
