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

const deleteDesc = `This command removes one or more versions of a chart from the repository.

'helm s3 delete' takes two arguments:
- NAME - name of the chart to delete,
- REPO - target repository.

Use --version once per version, or a comma-separated list, to remove several versions in a
single run (one index fetch and one index upload).

[Provenance]

If the chart is signed, the provenance file is removed from the repository as well.
`

const deleteExample = `  helm s3 delete epicservice --version 0.5.1 my-repo
  - removes version 0.5.1 of epicservice.

  helm s3 delete epicservice --version 0.5.1 --version 0.5.2 my-repo
  - removes both versions in one operation.

  helm s3 delete epicservice --version 0.5.1,0.5.2 my-repo
  - same as repeating --version for each value.`

func newDeleteCommand(opts *options) *cobra.Command {
	act := &deleteAction{
		printer:   nil,
		acl:       "",
		chartName: "",
		repoName:  "",
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
	flags.StringSliceVar(&act.versions, "version", nil, "Version(s) of the chart to delete. Repeat the flag or use comma-separated values.")
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

	versions []string
}

func (act *deleteAction) run(ctx context.Context) error {
	versions := expandVersions(act.versions)
	if len(versions) == 0 {
		return errors.New("at least one non-empty --version is required")
	}

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

	for _, ver := range versions {
		url, err := idx.Delete(act.chartName, ver)
		if err != nil {
			return err
		}

		if url != "" {
			if !strings.HasPrefix(url, repoEntry.URL()) {
				url = strings.TrimSuffix(repoEntry.URL(), "/") + "/" + url
			}

			if err := storage.DeleteChart(ctx, url); err != nil {
				return errors.WithMessage(err, "delete chart file from s3")
			}
		}
	}

	idx.UpdateGeneratedTime()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to s3")
	}

	if err := idx.WriteFile(repoEntry.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	if len(versions) == 1 {
		act.printer.Printf("Successfully deleted the chart from the repository.\n")
	} else {
		act.printer.Printf("Successfully deleted %d chart versions from the repository.\n", len(versions))
	}
	return nil
}

// expandVersions flattens comma-separated entries, trims space, drops empties, and dedupes
// while preserving first-seen order.
func expandVersions(in []string) []string {
	var out []string
	seen := make(map[string]struct{})
	for _, part := range in {
		for _, v := range strings.Split(part, ",") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
