// +build !windows

package lockfile

import "os"

func (l *LockFile) remove() error {
	err := os.Remove(l.filename)
	if err != nil {
		return err
	}

	return l.descriptor.Close()
}
