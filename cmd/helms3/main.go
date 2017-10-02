package main

import (
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envAWsDefaultRegion   = "AWS_DEFAULT_REGION"

	actionPush = "push"
	actionInit = "init"

	defaultTimeout = time.Second * 5
)

func main() {
	if len(os.Args) == 5 {
		runProxy(os.Args[4])
		return
	}

	initCmd := kingpin.Command(actionInit, "Initialize empty repository on AWS S3.")
	initURI := initCmd.Arg("uri", "URI of repository, e.g. s3://awesome-bucket/charts").
		Required().
		String()
	pushCmd := kingpin.Command(actionPush, "Push chart to repository.")
	pushChartPath := pushCmd.Arg("chartPath", "Path to a chart, e.g. ./epicservice-0.5.1.tgz").
		Required().
		String()
	pushTargetRepository := pushCmd.Arg("repo", "Target repository to runPush").
		Required().
		String()
	action := kingpin.Parse()

	switch action {

	case actionInit:
		runInit(*initURI)
		return

	case actionPush:
		runPush(*pushChartPath, *pushTargetRepository)
		return

	}
}
