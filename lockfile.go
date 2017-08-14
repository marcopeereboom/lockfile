// lockfile - package that manages filesystem locks.
//
// This package enables the caller to use filesystem locks.  It uses the
// operating system file creation system call with the exclusive flag.
//
// This package relies on callers to always check errors.  Failure to do so may
// result in corruption.
package lockfile

import (
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	ErrTimeout = errors.New("timeout")
)

// LockFile opaque type that contains the lockfile context.
type LockFile struct {
	dirname    string
	resolution time.Duration
}

// New returns a LockFile context if the directory is writable.
func New(dirname string, resolution time.Duration) (*LockFile, error) {
	l := LockFile{
		dirname:    dirname,
		resolution: resolution,
	}

	err := os.Mkdir(dirname, 0600)
	if err != nil {
		return nil, err
	}

	err = os.Remove(dirname)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// Resolution sets the retry resolution.  The default retry resolution is 100ms.
func (l *LockFile) Resolution(resolution time.Duration) {
	l.resolution = resolution
}

func (l *LockFile) TryLock() bool {
	err := os.Mkdir(l.dirname, 0600)
	if err == nil {
		return true
	}
	return false
}

// Lock attempts to create a lock file within the given timeout.  If the lock
// can not be obtained within the given timeout or due to a filesystem error it
// returns failure.  The caller must check the error.
func (l *LockFile) Lock(timeout time.Duration) error {
	end := time.Now().Add(timeout)
	for {
		if l.TryLock() {
			return nil
		}
		if time.Now().Before(end) {
			time.Sleep(l.resolution)
			continue
		}
		return ErrTimeout
	}

	return fmt.Errorf("not reached")
}

// Unlock attempts to unlock by removing the lock file.  The caller must check
// the error.
func (l *LockFile) Unlock() error {
	return os.Remove(l.dirname)
}
