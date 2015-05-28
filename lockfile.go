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
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

var (
	ErrTimeout   = errors.New("timeout")
	ErrNotLocked = errors.New("not locked")
	ErrLocked    = errors.New("locked")
)

const (
	resolution = 100 * time.Millisecond
)

// LockFile opaque type that contains the lockfile context.
type LockFile struct {
	mtx        sync.Mutex
	filename   string
	resolution time.Duration
	descriptor *os.File
}

// New returns a LockFile context if the directory is writable.
func New(filename string) (*LockFile, error) {
	l := LockFile{
		filename:   filename,
		resolution: resolution,
	}

	// make sure we can write to directory
	fd, err := ioutil.TempFile(path.Dir(filename), "testlock")
	if err != nil {
		return nil, err
	}
	err = os.Remove(fd.Name())
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// Resolution sets the retry resolution.  The default retry resolution is 100ms.
func (l *LockFile) Resolution(resolution time.Duration) {
	l.resolution = resolution
}

// Lock attempts to create a lock file within the given timeout.  If the lock
// can not be obtained within the given timeout or due to a filesystem error it
// returns failure.  The caller must check the error.
func (l *LockFile) Lock(timeout time.Duration) error {
	var err error

	end := time.Now().Add(timeout)
	for {
		l.mtx.Lock()
		if l.descriptor != nil {
			return ErrLocked
		}
		l.descriptor, err = os.OpenFile(l.filename,
			os.O_CREATE|os.O_EXCL, 0600)
		l.mtx.Unlock()

		if os.IsExist(err) {
			if time.Now().Before(end) {
				time.Sleep(resolution)
				continue
			}
			return ErrTimeout
		} else if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("not reached")
}

// Unlock attempts to unlock by removing the lock file.  The caller must check
// the error.
func (l *LockFile) Unlock() error {
	var err error

	l.mtx.Lock()
	defer l.mtx.Unlock()

	if l.descriptor == nil {
		return ErrNotLocked
	}

	defer func() {
		l.descriptor = nil
	}()

	err = os.Remove(l.filename)
	if err != nil {
		return err
	}

	return l.descriptor.Close()
}
