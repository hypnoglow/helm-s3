package main

import (
	"context"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	"github.com/hypnoglow/helm-s3/internal/locks"
	"github.com/spf13/cobra"
)

const lockDesc = `This command creates or removes dynamodb locks for a repository.

'helm s3 lock' takes one argument:
- REPO - target repository.
`

const lockExample = `  helm s3 lock my-repo - creates a dyanmodb lock for 'my-repo'.
  helm s3 lock --unlock my-repo - removes the dynamodb lock for 'my-repo'`

func newLockCommand(opts *options) *cobra.Command {
	act := &lockAction{}

	cmd := &cobra.Command{
		Use:     "lock REPO",
		Short:   "Lock or unlock the given repository.",
		Long:    lockDesc,
		Example: lockExample,
		Args:    wrapPositionalArgsBadUsage(cobra.ExactArgs(1)),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// No completions for the NAME and REPO arguments.
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repoEntry, err := helmutil.LookupRepoEntry(args[0])
			if err != nil {
				return err
			}

			act.lock, err = locks.NewDynamoDBLockWithDefaultConfig(opts.dynamodbLockTableName)
			if err != nil {
				return err
			}

			act.lockID = awss3.LockID(repoEntry.URL())
			return act.run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&act.unlock, "unlock", act.unlock, "If provided unlock the lock instead of acquiring it.")
	_ = cobra.MarkFlagRequired(flags, "version")

	return cmd
}

type lockAction struct {
	lock   locks.Lock
	unlock bool
	lockID string
}

func (act *lockAction) run(ctx context.Context) error {
	if act.unlock {
		return act.lock.Unlock(ctx, act.lockID)
	}

	return act.lock.Lock(ctx, act.lockID)
}
