package locks

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAlreadyLocked = errors.New("lock is already acquired")
	ErrLockTimeout   = errors.New("timed out waiting for lock")
)

type Lock interface {
	Lock(ctx context.Context, lockID string) error
	Unlock(ctx context.Context, lockID string) error
}

// IsLocked checks if the Lock is locked for the given lockID
func IsLocked(ctx context.Context, lock Lock, lockID string) bool {
	err := lock.Lock(ctx, lockID)
	// Skip unlocking if ErrAlreadyLocked since we don't want to accidentally
	// release the lock.
	if errors.Is(err, ErrAlreadyLocked) {
		return true
	}

	// We only want to unlock if the lock was not already locked.
	defer lock.Unlock(ctx, lockID)
	return false
}

func WaitForLock(ctx context.Context, lock Lock, lockID string, lockTimeoutSeconds int) error {
	for start := time.Now(); time.Since(start) < time.Second*time.Duration(lockTimeoutSeconds); {
		select {
		case <-ctx.Done():
			return ErrLockTimeout
		default:
			err := lock.Lock(ctx, lockID)
			if err != nil && errors.Is(err, ErrAlreadyLocked) {
				time.Sleep(time.Second)
				continue
			}

			return err
		}
	}

	return ErrLockTimeout
}
