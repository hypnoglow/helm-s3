package locks

import "context"

// FalseLock implements a no-op version of the locks.Lock interface
type FalseLock struct{}

func NewFalseLock() *FalseLock {
	return &FalseLock{}
}

func (fl *FalseLock) Lock(ctx context.Context, lockID string) error {
	return nil
}

func (fl *FalseLock) Unlock(ctx context.Context, lockID string) error {
	return nil
}
