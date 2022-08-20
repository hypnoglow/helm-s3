package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const initDesc = `This command initializes an empty repository on AWS S3.

'helm s3 init' takes one argument:
- URI - URI of the repository.
`

const initExample = `  helm s3 init s3://awesome-bucket/charts - inits chart repository in 'awesome-bucket' bucket under 'charts' path.`

func newInitCommand(opts *options) *cobra.Command {
	act := &initAction{
		printer: nil,
		acl:     "",
		uri:     "",
	}

	cmd := &cobra.Command{
		Use:     "init URI",
		Short:   "Initialize empty repository on AWS S3.",
		Long:    initDesc,
		Example: initExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(1)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the URI argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.acl = opts.acl
			act.uri = args[0]
			return act.run(cmd.Context())
		},
	}

	return cmd
}

type initAction struct {
	printer printer

	// global flags

	acl string

	// args

	uri string
}

func (act *initAction) run(ctx context.Context) error {
	r, err := helmutil.NewIndex().Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	sess, err := awsutil.Session(awsutil.DynamicBucketRegion(act.uri))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	if err := storage.PutIndex(ctx, act.uri, act.acl, r); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	// TODO:
	// do we need to automatically do `helm repo add <name> <uri>`,
	// like we are doing `helm repo update` when we push a chart
	// with this plugin?

	act.printer.Printf("Initialized empty repository at %s\n", act.uri)
	return nil
}
