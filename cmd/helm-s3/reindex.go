package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const reindexDesc = `This command performs a reindex of the repository.

'helm s3 push' takes one argument:
- REPO - target repository.
`

const reindexExample = `  helm s3 reindex my-repo - performs a reindex of the repository with name 'my-repo'.`

func newReindexCommand(opts *options) *cobra.Command {
	act := &reindexAction{
		printer:  nil,
		acl:      "",
		verbose:  false,
		repoName: "",
		relative: false,
	}

	cmd := &cobra.Command{
		Use:     "reindex REPO",
		Short:   "Reindex the repository.",
		Long:    reindexDesc,
		Example: reindexExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(1)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the REPO argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.acl = opts.acl
			act.verbose = opts.verbose
			act.repoName = args[0]
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&act.relative, "relative", act.relative, "Use relative chart URLs in the index instead of absolute.")

	return cmd
}

type reindexAction struct {
	printer printer

	// global flags

	acl     string
	verbose bool

	// args

	repoName string

	// flags

	relative bool
}

func (act *reindexAction) run(ctx context.Context) error {
	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	sess, err := awsutil.Session(awsutil.DynamicBucketRegion(repoEntry.URL()))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	items, errs := storage.Traverse(ctx, repoEntry.URL())

	builtIndex := make(chan helmutil.Index, 1)
	go func() {
		idx := helmutil.NewIndex()
		for item := range items {
			baseURL := repoEntry.URL()
			if act.relative {
				baseURL = ""
			}

			if act.verbose {
				act.printer.Printf("[DEBUG] Adding %s to index.\n", item.Filename)
			}

			filename := escapeIfRelative(item.Filename, act.relative)

			if err := idx.Add(item.Meta.Value(), filename, baseURL, item.Hash); err != nil {
				act.printer.PrintErrf("[ERROR] failed to add chart to the index: %s", err)
			}
		}
		idx.SortEntries()
		idx.UpdateGeneratedTime()

		builtIndex <- idx
	}()

	for err = range errs {
		return fmt.Errorf("traverse the chart repository: %v", err)
	}

	idx := <-builtIndex

	r, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, r); err != nil {
		return errors.Wrap(err, "upload index to the repository")
	}

	if err := idx.WriteFile(repoEntry.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	act.printer.Printf("Repository %s was successfully reindexed.\n", act.repoName)
	return nil
}
