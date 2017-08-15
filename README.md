# lockfile - package that manages filesystem locks.

The lockfile package uses the filesystem to serialize cooperative processes.
Additionally it uses an internal mutex to provide single process serialization.

Typical usage:
```Go
	// Create lock
	l, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		return err
	}

	...

	// Lock something
	err = l.Lock(5*time.Second)
	if err != nil {
		return err
	}

	...

	// And unlock it
	err = l.Unlock()
	if err != nil {
		return err
	}
```
