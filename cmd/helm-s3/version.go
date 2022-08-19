package main

import (
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const versionDesc = `This command prints plugin version.

You can also check the Helm mode in which the plugin operates, either v2 or v3.
If the plugin does not detect Helm version properly, you can forcefully change 
the mode: set HELM_S3_MODE environment variable to either 2 or 3.
`

func newVersionCommand() *cobra.Command {
	act := &versionAction{
		mode: false,
	}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print plugin version.",
		Long:    versionDesc,
		Example: "",
		Args:    wrapPositionalArgsBadUsage(cobra.NoArgs),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			return act.run()
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&act.mode, "mode", act.mode, "Also print Helm version mode in which the plugin operates, either v2 or v3.")

	return cmd
}

type versionAction struct {
	printer printer

	// flags

	mode bool
}

func (act *versionAction) run() error {
	if !act.mode {
		act.printer.Printf("%v\n", version)
		return nil
	}

	mode := "v2"
	if helmutil.IsHelm3() {
		mode = "v3"
	}

	act.printer.Printf("helm-s3 plugin version: %s\n", version)
	act.printer.Printf("Helm version mode: %s\n", mode)
	return nil
}
