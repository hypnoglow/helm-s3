package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
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

	defaultTimeout = time.Second * 5
)

func main() {
	if len(os.Args) == 5 {
		if err := runProxy(os.Args[4]); err != nil {
			log.Fatal(err)
		}
		return
	}

	cli := kingpin.New("helm s3", "")
	cli.Command(actionVersion, "Show plugin version.")

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

	reindexCmd := cli.Command(actionReindex, "Reindex the repository.")
	reindexTargetRepository := reindexCmd.Arg("repo", "Target repository to reindex").
		Required().
		String()

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

	switch action {
	case actionVersion:
		fmt.Print(version)
		return

	case actionInit:
		if err := runInit(*initURI); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Initialized empty repository at %s\n", *initURI)
		return

	case actionPush:
		if err := runPush(*pushChartPath, *pushTargetRepository); err != nil {
			log.Fatal(err)
		}
		return

	case actionReindex:
		if err := runReindex(*reindexTargetRepository); err != nil {
			log.Fatal(err)
		}

	case actionDelete:
		if err := runDelete(*deleteChartName, *deleteChartVersion, *deleteTargetRepository); err != nil {
			log.Fatal(err)
		}
		return
	}
}
