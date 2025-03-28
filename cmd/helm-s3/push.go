package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const pushDesc = `This command uploads a chart to the repository.

'helm s3 push' takes two arguments:
- PATH - path to the chart file,
- REPO - target repository.

[Provenance]

If the chart is signed, the provenance file is uploaded to the repository as well.
`

const pushExample = `  helm s3 push ./epicservice-0.5.1.tgz my-repo - uploads chart file 'epicservice-0.5.1.tgz' from the current directory to the repository with name 'my-repo'.`

func newPushCommand(opts *options) *cobra.Command {
	contentTypeDefault := os.Getenv("S3_CHART_CONTENT_TYPE")
	if contentTypeDefault == "" {
		contentTypeDefault = "application/gzip"
	}

	act := &pushAction{
		printer:        nil,
		acl:            "",
		chartPath:      "",
		repoName:       "",
		contentType:    contentTypeDefault,
		dryRun:         false,
		force:          false,
		ignoreIfExists: false,
		relative:       false,
	}

	cmd := &cobra.Command{
		Use:     "push PATH REPO",
		Short:   "Push chart to the repository.",
		Long:    pushDesc,
		Example: pushExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(2)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				// Allow file completion for the PATH argument.
				return nil, cobra.ShellCompDirectiveDefault
			}
			// No completions for the REPO argument.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			act.printer = cmd
			act.acl = opts.acl
			act.chartPath = args[0]
			act.repoName = args[1]
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&act.contentType, "content-type", act.contentType, "Set the content-type for the chart file. Can be sourced from S3_CHART_CONTENT_TYPE environment variable.")
	flags.BoolVar(&act.dryRun, "dry-run", act.dryRun, "Simulate push operation, but don't actually touch anything.")
	flags.BoolVar(&act.force, "force", act.force, "Replace the chart if it already exists. This can cause the repository to lose existing chart; use it with care.")
	flags.BoolVar(&act.ignoreIfExists, "ignore-if-exists", act.ignoreIfExists, "If the chart already exists, exit normally and do not trigger an error.")
	flags.BoolVar(&act.relative, "relative", act.relative, "Use relative chart URL in the index instead of absolute.")

	// We don't use cobra's feature
	//
	//  cmd.MarkFlagsMutuallyExclusive("force", "ignore-if-exists")
	//
	// because the error message is confusing. Instead, we check the flags
	// manually in run().

	return cmd
}

type pushAction struct {
	printer printer

	// global args

	acl string

	// args

	chartPath string
	repoName  string

	// flags

	contentType    string
	dryRun         bool
	force          bool
	ignoreIfExists bool
	relative       bool
}

func (act *pushAction) run(ctx context.Context) error { //nolint:gocyclo // Maybe refactor later.
	// Sanity check.
	if act.force && act.ignoreIfExists {
		act.printer.PrintErrf(
			"The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together.\n",
		)
		return newSilentError()
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

	fpath, err := filepath.Abs(act.chartPath)
	if err != nil {
		return errors.WithMessage(err, "get chart abs path")
	}

	dir := filepath.Dir(fpath)
	fname := filepath.Base(fpath)

	if err := os.Chdir(dir); err != nil {
		return errors.Wrapf(err, "change dir to %s", dir)
	}

	// Load chart, calculate required params like hash,
	// and upload the chart right away.

	chart, err := helmutil.LoadChart(fname)
	if err != nil {
		return err
	}

	if cachedIndex, err := helmutil.LoadIndex(repoEntry.CacheFile()); err == nil {
		// if cached index exists, check if the same chart version exists in it.
		if cachedIndex.Has(chart.Name(), chart.Version()) {
			if act.ignoreIfExists {
				return act.ignoreIfExistsError()
			}
			if !act.force {
				return act.chartExistsError()
			}

			// fallthrough on --force
		}
	}

	hash, err := helmutil.DigestFile(fname)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

	chartFile, err := os.Open(fname)
	if err != nil {
		return errors.Wrap(err, "open chart file")
	}

	hasProv := false
	provFile, err := os.Open(fname + ".prov")
	switch {
	case err == nil:
		hasProv = true
	case errors.Is(err, os.ErrNotExist):
		// No provenance file, ignore it.
	case err != nil:
		return fmt.Errorf("open prov file: %w", err)
	}

	exists, err := storage.Exists(ctx, repoEntry.URL()+"/"+fname)
	if err != nil {
		return errors.WithMessage(err, "check if chart already exists in the repository")
	}

	if exists {
		if act.ignoreIfExists {
			return act.ignoreIfExistsError()
		}
		if !act.force {
			return act.chartExistsError()
		}

		// fallthrough on --force
	}

	if !act.dryRun {
		chartMetaJSON, err := chart.Metadata().MarshalJSON()
		if err != nil {
			return err
		}
		if _, err := storage.PutChart(
			ctx,
			repoEntry.URL()+"/"+fname,
			chartFile,
			string(chartMetaJSON),
			act.acl,
			hash,
			act.contentType,
			hasProv,
			provFile,
		); err != nil {
			return errors.WithMessage(err, "upload chart to s3")
		}
	}

	// The gap between index fetching and uploading should be as small as
	// possible to make the best effort to avoid race conditions.
	// See https://github.com/hypnoglow/helm-s3/issues/18 for more info.

	// Fetch current index, update it and upload it back.

	b, err := storage.FetchRaw(ctx, repoEntry.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	baseURL := repoEntry.URL()
	if act.relative {
		baseURL = ""
	}

	filename := escapeIfRelative(fname, act.relative)

	if err := idx.AddOrReplace(chart.Metadata().Value(), filename, baseURL, hash); err != nil {
		return errors.WithMessage(err, "add/replace chart in the index")
	}
	idx.SortEntries()
	idx.UpdateGeneratedTime()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	if !act.dryRun {
		if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, idxReader); err != nil {
			return errors.WithMessage(err, "upload index to s3")
		}

		if err := idx.WriteFile(repoEntry.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
			return errors.WithMessage(err, "update local index")
		}
	}

	act.printer.Printf("Successfully uploaded the chart to the repository.\n")
	return nil
}

func (act *pushAction) ignoreIfExistsError() error {
	act.printer.Printf(
		"The chart already exists in the repository, keep existing chart and ignore push.\n",
	)
	return nil
}

func (act *pushAction) chartExistsError() error {
	act.printer.PrintErrf(
		"The chart already exists in the repository and cannot be overwritten without an explicit intent.\n\n"+
			"If you want to replace existing chart, use --force flag:\n\n"+
			"  helm s3 push --force %[1]s %[2]s\n\n"+
			"If you want to ignore this error, use --ignore-if-exists flag:\n\n"+
			"  helm s3 push --ignore-if-exists %[1]s %[2]s\n\n",
		act.chartPath,
		act.repoName,
	)
	return newSilentError()
}
