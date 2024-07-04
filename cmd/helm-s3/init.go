package main

import (
	"context"
	"fmt"
	"io/fs"

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
		printer:        nil,
		acl:            "",
		uri:            "",
		force:          false,
		ignoreIfExists: false,
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

	flags := cmd.Flags()
	flags.BoolVar(&act.force, "force", act.force, "Replace the index file if it already exists.")
	flags.BoolVar(&act.ignoreIfExists, "ignore-if-exists", act.ignoreIfExists, "If the index file already exists, exit normally and do not trigger an error.")

	// We don't use cobra's feature
	//
	//  cmd.MarkFlagsMutuallyExclusive("force", "ignore-if-exists")
	//
	// because the error message is confusing. Instead, we check the flags
	// manually in run().

	return cmd
}

type initAction struct {
	printer printer

	// global flags

	acl string

	// args

	uri string

	// flags

	force          bool
	ignoreIfExists bool
}

func (act *initAction) run(ctx context.Context) error {
	if act.force && act.ignoreIfExists {
		act.printer.PrintErrf(
			"The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together.\n",
		)
		return newSilentError()
	}

	if err := act.checkRepoEntry(); err != nil {
		return err
	}

	r, err := helmutil.NewIndex().Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	sess, err := awsutil.Session(awsutil.DynamicBucketRegion(act.uri))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	exists, err := storage.IndexExists(ctx, act.uri)
	if err != nil {
		return fmt.Errorf("check if index exists in the storage: %v", err)
	}
	if exists {
		if act.ignoreIfExists {
			return act.ignoreIfExistsInStorageError()
		}
		if !act.force {
			return act.alreadyExistsInStorageError()
		}

		// fallthrough on --force
	}

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

func (act *initAction) checkRepoEntry() error {
	repoEntry, ok, err := helmutil.LookupRepoEntryByURL(act.uri)
	if errors.Is(err, fs.ErrNotExist) {
		// Repo file may not exist, this is OK for instance when the helm is
		// just installed (e.g. in docker).
		return nil
	}
	if err != nil {
		return fmt.Errorf("lookup repo entry by url: %v", err)
	}

	if !ok {
		// Repo entry not found - all is good.
		return nil
	}

	// Repo entry exists.

	if act.ignoreIfExists {
		return act.ignoreIfExistsError(repoEntry.Name())
	}
	if !act.force {
		return act.alreadyExistsError(repoEntry.Name())
	}

	// fallthrough on --force
	return nil
}

func (act *initAction) ignoreIfExistsError(name string) error {
	act.printer.Printf(
		"The repository with the provided URI already exists under name %q, ignore init operation.\n",
		name,
	)
	return nil
}

func (act *initAction) ignoreIfExistsInStorageError() error {
	act.printer.Printf(
		"The index file already exists under the provided URI, ignore init operation.\n",
	)
	return nil
}

func (act *initAction) alreadyExistsError(name string) error {
	act.printer.PrintErrf(
		"The repository with the provided URI already exists under name %[1]q, the index file and cannot be overwritten without an explicit intent.\n\n"+
			"If you want to replace existing index file, use --force flag:\n\n"+
			"  helm s3 init --force %[2]s\n\n"+
			"If you want to ignore this error, use --ignore-if-exists flag:\n\n"+
			"  helm s3 init --ignore-if-exists %[2]s\n\n",
		name,
		act.uri,
	)
	return newSilentError()
}

func (act *initAction) alreadyExistsInStorageError() error {
	act.printer.PrintErrf(
		"The index file already exists under the provided URI and cannot be overwritten without an explicit intent.\n\n"+
			"If you want to replace existing index file, use --force flag:\n\n"+
			"  helm s3 init --force %[1]s\n\n"+
			"If you want to ignore this error, use --ignore-if-exists flag:\n\n"+
			"  helm s3 init --ignore-if-exists %[1]s\n\n",
		act.uri,
	)
	return newSilentError()
}
