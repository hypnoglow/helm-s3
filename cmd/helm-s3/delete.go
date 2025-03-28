package main

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const deleteDesc = `This command removes a chart from the repository.

'helm s3 delete' takes two arguments:
- NAME - name of the chart to delete,
- REPO - target repository.

[Provenance]

If the chart is signed, the provenance file is removed from the repository as well.
`

const deleteExample = `  helm s3 delete epicservice --version 0.5.1 my-repo - removes the chart with name 'epicservice' and version 0.5.1 from the repository with name 'my-repo'.`

func newDeleteCommand(opts *options) *cobra.Command {
	act := &deleteAction{
		printer:   nil,
		acl:       "",
		chartName: "",
		repoName:  "",
		version:   "",
	}

	cmd := &cobra.Command{
		Use:     "delete NAME REPO",
		Aliases: []string{"del"},
		Short:   "Delete chart from the repository.",
		Long:    deleteDesc,
		Example: deleteExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(2)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the NAME and REPO arguments.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.acl = opts.acl
			act.chartName = args[0]
			act.repoName = args[1]
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&act.version, "version", act.version, "Version of the chart to delete.")
	_ = cobra.MarkFlagRequired(flags, "version")

	return cmd
}

type deleteAction struct {
	printer printer

	// global flags

	acl string

	// args

	chartName string
	repoName  string

	// flags

	version string
}

func (act *deleteAction) run(ctx context.Context) error {
	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	sess, err := awsutil.Session(awsutil.DynamicBucketRegion(repoEntry.URL()))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	// Fetch current index.
	b, err := storage.FetchRaw(ctx, repoEntry.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Update index.

	url, err := idx.Delete(act.chartName, act.version)
	if err != nil {
		return err
	}
	idx.UpdateGeneratedTime()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	// Delete the file from S3 and replace index file.

	if url != "" {
		// For relative URLs we need to prepend base URL.
		if !strings.HasPrefix(url, repoEntry.URL()) {
			url = strings.TrimSuffix(repoEntry.URL(), "/") + "/" + url
		}

		if err := storage.DeleteChart(ctx, url); err != nil {
			return errors.WithMessage(err, "delete chart file from s3")
		}
	}

	if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to s3")
	}

	if err := idx.WriteFile(repoEntry.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	act.printer.Printf("Successfully deleted the chart from the repository.\n")
	return nil
}
