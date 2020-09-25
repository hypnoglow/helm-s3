package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

var (
	version = "master"
)

const (
	actionVersion = "version"
	actionInit    = "init"
	actionPush    = "push"
	actionReindex = "reindex"
	actionDelete  = "delete"

	defaultTimeout       = time.Minute * 5
	defaultTimeoutString = "5m"

	// duplicated in e2e package for testing.
	defaultChartsContentType = "application/gzip"

	helpFlagTimeout = `Timeout for the whole operation to complete. Defaults to 5 minutes.

If you don't use MFA, it may be reasonable to lower the timeout
for the most commands, for example to 10 seconds.

In opposite, in cases where you want to reindex big repository
(e.g. 10 000 charts), you definitely want to increase the timeout.
`

	helpFlagACL = `S3 Object ACL to set on the Chart and Index object.

For more information on S3 ACLs please see https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl
`
	relativeFlag     = "relative"
	helpRelativeFlag = "Index using relative URLs (useful when S3 buckets are replicated)"
)

// Action describes plugin action that can be run.
type Action interface {
	Run(context.Context) error
}

func main() {
	helmutil.SetupHelm()

	log.SetFlags(0)

	if len(os.Args) == 5 && !isAction(os.Args[1]) {
		cmd := proxyCmd{uri: os.Args[4]}
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		if err := cmd.Run(ctx); err != nil {
			log.Fatal(err)
		}
		return
	}

	cli := kingpin.New("helm s3", "")
	versionCmd := cli.Command(actionVersion, "Show plugin version.")
	versionMode := versionCmd.Flag("mode", "Also print Helm version mode in which the plugin operates, either v2 or v3. In case when the plugin does not detect Helm version properly, you can forcefully change the mode: set HELM_S3_MODE environment variable to either 2 or 3.").
		Bool()

	timeout := cli.Flag("timeout", helpFlagTimeout).
		Default(defaultTimeoutString).
		Duration()

	acl := cli.Flag("acl", helpFlagACL).
		Default("").
		OverrideDefaultFromEnvar("S3_ACL").
		String()

	initCmd := cli.Command(actionInit, "Initialize empty repository on AWS S3.")
	initURI := initCmd.Arg("uri", "URI of repository, e.g. s3://awesome-bucket/charts").
		Required().
		String()

	pushCmd := cli.Command(actionPush, "Push chart to the repository.")
	pushChartPath := pushCmd.Arg("chartPath", "Path to a chart, e.g. ./epicservice-0.5.1.tgz").
		Required().
		String()
	pushTargetRepository := pushCmd.Arg("repo", "Target repository to push to").
		Required().
		String()
	pushForce := pushCmd.Flag("force", "Replace the chart if it already exists. This can cause the repository to lose existing chart; use it with care.").
		Bool()
	pushDryRun := pushCmd.Flag("dry-run", "Simulate a push, but don't actually touch anything.").
		Bool()
	pushIgnoreIfExists := pushCmd.Flag("ignore-if-exists", "If the chart already exists, exit normally and do not trigger an error.").
		Bool()
	pushContentType := pushCmd.Flag("content-type", "Set the Charts content-type").
		Default(defaultChartsContentType).
		OverrideDefaultFromEnvar("S3_CHART_CONTENT_TYPE").
		String()
	pushRelative := pushCmd.Flag(relativeFlag, helpRelativeFlag).Bool()

	reindexCmd := cli.Command(actionReindex, "Reindex the repository.")
	reindexTargetRepository := reindexCmd.Arg("repo", "Target repository to reindex").
		Required().
		String()
	reindexRelative := reindexCmd.Flag(relativeFlag, helpRelativeFlag).Bool()

	deleteCmd := cli.Command(actionDelete, "Delete chart from the repository.").Alias("del")
	deleteChartName := deleteCmd.Arg("chartName", "Name of chart to delete").
		Required().
		String()
	deleteChartVersion := deleteCmd.Flag("version", "Version of chart to delete").
		Required().
		String()
	deleteTargetRepository := deleteCmd.Arg("repo", "Target repository to delete from").
		Required().
		String()

	action := kingpin.MustParse(cli.Parse(os.Args[1:]))
	if action == "" {
		cli.Usage(os.Args[1:])
		os.Exit(0)
	}

	var act Action
	switch action {
	case actionVersion:
		if *versionMode {
			mode := "v2"
			if helmutil.IsHelm3() {
				mode = "v3"
			}
			fmt.Printf("helm-s3 plugin version: %s\n", version)
			fmt.Printf("Helm version mode: %s\n", mode)
			return
		}
		fmt.Print(version)
		return

	case actionInit:
		act = initAction{
			uri: *initURI,
			acl: *acl,
		}
		defer fmt.Printf("Initialized empty repository at %s\n", *initURI)

	case actionPush:
		act = pushAction{
			chartPath:      *pushChartPath,
			repoName:       *pushTargetRepository,
			force:          *pushForce,
			dryRun:         *pushDryRun,
			ignoreIfExists: *pushIgnoreIfExists,
			acl:            *acl,
			contentType:    *pushContentType,
			relative:       *pushRelative,
		}

	case actionReindex:
		act = reindexAction{
			repoName: *reindexTargetRepository,
			acl:      *acl,
			relative: *reindexRelative,
		}
		defer fmt.Printf("Repository %s was successfully reindexed.\n", *reindexTargetRepository)

	case actionDelete:
		act = deleteAction{
			name:     *deleteChartName,
			version:  *deleteChartVersion,
			repoName: *deleteTargetRepository,
			acl:      *acl,
		}
	default:
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	err := act.Run(ctx)
	switch err {
	case nil:
	case ErrChartExists:
		log.Fatalf("The chart already exists in the repository and cannot be overwritten without an explicit intent. If you want to replace existing chart, use --force flag:\n\n\thelm s3 push --force %s %s\n\n", *pushChartPath, *pushTargetRepository)
	default:
		log.Fatal(err)
	}
}

func isAction(name string) bool {
	return name == actionDelete ||
		name == actionInit ||
		name == actionPush ||
		name == actionReindex ||
		name == actionVersion
}
