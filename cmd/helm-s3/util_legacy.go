package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// detectDownload detects if the plugin runs in downloader mode.
//
// This function is used for compatibility with older implementation
// before the plugin CLI moved to cobra module.
// Remove before v1 release.
func detectDownload(root *cobra.Command, args []string) bool {
	// Downloader plugins has exactly 5 arguments.
	// See https://helm.sh/docs/topics/plugins/#downloader-plugins

	const numArgs = 5
	if len(args) != numArgs {
		return false
	}

	possibleCommand := args[1]
	if isKnownCommand(root.Commands(), possibleCommand) {
		return false
	}

	act := &downloadAction{
		printer:  root,
		certFile: args[1],
		keyFile:  args[2],
		caFile:   args[3],
		url:      args[4],
	}

	const defaultTimeout = 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	err := act.run(ctx)
	cancel()
	if err != nil {
		root.Print(err)
		os.Exit(1)
	}

	return true
}

func isKnownCommand(commands []*cobra.Command, name string) bool {
	for _, cmd := range commands {
		if name == cmd.Name() {
			return true
		}
	}

	return false
}
