package lockfile

import "os"

func (l *LockFile) remove() error {
	// in windows it should be ok to flip order of remove close
	err := l.descriptor.Close()
	if err != nil {
		return err
	}

	return os.Remove(l.filename)
}
