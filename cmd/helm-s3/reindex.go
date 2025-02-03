package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

const reindexDesc = `This command performs a reindex of the repository.

'helm s3 push' takes one argument:
- REPO - target repository.
`

const reindexExample = `  helm s3 reindex my-repo - performs a reindex of the repository with name 'my-repo'.`
const batchSize = 1000

func newReindexCommand(opts *options) *cobra.Command {
	act := &reindexAction{
		printer:  nil,
		acl:      "",
		verbose:  false,
		repoName: "",
		relative: false,
		dryRun:   false,
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
	flags.BoolVar(&act.dryRun, "dry-run", act.dryRun, "Simulate reindex, don't push it to the dest repo.")

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

	dryRun bool
}

func (act *reindexAction) run(ctx context.Context) error {
	start := time.Now()
	log.Infof("Starting reindex for %s", start)
	act.printer.Printf("[DEBUG] Starting reindex.\n")

	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	sess, err := awsutil.Session(awsutil.DynamicBucketRegion(repoEntry.URL()))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	items, _ := storage.Traverse(ctx, repoEntry.URL())

	// Use a buffered channel to handle the concurrent indexing
	builtIndex := make(chan *repo.IndexFile, len(items)/batchSize+1)
	var wg sync.WaitGroup

	act.printer.Printf("[DEBUG] Creating split indexes.\n")

	// Process items in batches of 1000
	log.Info("Processing items in batches of 1000")
	log.Info("Total items: ", len(items))
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		wg.Add(1)
		go func(batch []awss3.ChartInfo) {
			defer wg.Done()
			idx := repo.NewIndexFile()

			for _, item := range batch {
				baseURL := repoEntry.URL()
				if act.relative {
					baseURL = ""
				}

				if act.verbose {
					act.printer.Printf("[DEBUG] Adding %s to index.\n", item.Filename)
				}

				filename := escapeIfRelative(item.Filename, act.relative)

				if err := idx.MustAdd(item.Meta.Value().(*chart.Metadata), filename, baseURL, item.Hash); err != nil {
					act.printer.PrintErrf("[ERROR] failed to add chart to the index: %s", err)
				}
			}

			builtIndex <- idx
		}(items[i:end])
	}

	log.Info("Waiting for all goroutines to finish")
	// Wait for all goroutines to finish
	wg.Wait()
	close(builtIndex)

	log.Info("Processing indexes")
	// Merge the individual index files into a single index file
	finalIndex := repo.NewIndexFile()
	for idx := range builtIndex {
		finalIndex.Merge(idx)
	}

	finalIndex.SortEntries()

	if err := finalIndex.WriteFile(repoEntry.CacheFile(), helmutil.DefaultIndexFilePerm); err != nil {
		return errors.WithMessage(err, "update local index")
	}
	log.Infof("Index file written to %s", repoEntry.CacheFile())

	file, err := os.Open(repoEntry.CacheFile())
	if err != nil {
		return errors.Wrap(err, "open index file")
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, "get file stats")
	}

	// Read the file into a byte slice
	ra := make([]byte, stat.Size())
	if _, err := bufio.NewReader(file).Read(ra); err != nil && err != io.EOF {
		return errors.Wrap(err, "read index file")
	}

	r := bytes.NewReader(ra)

	if !act.dryRun {
		if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, r); err != nil {
			return errors.Wrap(err, "upload index to the repository")
		}
	} else {
		act.printer.Printf("[DEBUG] Dry run, not pushing index to the repository.\n")
	}

	act.printer.Printf("Repository %s was successfully reindexed.\n", act.repoName)
	log.Infof("Reindex done in %s", time.Since(start))
	return nil
}
