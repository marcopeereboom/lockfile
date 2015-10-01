package lockfile

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	lockfile string
)

func init() {
	// best effort
	fd, err := ioutil.TempFile("", "testlock")
	if err != nil {
		panic(err)
	}
	fd.Close()

	lockfile = fd.Name()
	err = os.Remove(fd.Name())
	if err != nil {
		panic(err)
	}
}

func TestRace(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = l1.Lock(time.Second)
			if err != nil {
				t.Fatalf("l1 %v: %v", lockfile, err)
			}

			err = l1.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = l2.Lock(time.Second)
			if err != nil {
				t.Fatalf("l2 %v: %v", lockfile, err)
			}

			err = l2.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		}()
		wg.Wait()
	}
}

func TestLockUnlockRace(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan error)

	for i := 0; i < 10; i++ {
		err = l1.Lock(time.Second)
		if err != nil {
			t.Fatalf("l1 %v: %v", lockfile, err)
		}

		go func() {
			c <- l2.Lock(time.Second)
		}()

		// unlock l1
		err = l1.Unlock()
		if err != nil {
			t.Fatal(err)
		}

		err = <-c
		if err != nil {
			t.Fatal(err)
		}

		// unlock l2
		err = l2.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestLockUnlockBefore(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatalf("l1 %v: %v", lockfile, err)
	}

	c := make(chan error)
	go func() {
		time.Sleep(250 * time.Millisecond)
		c <- l2.Lock(time.Second)
	}()

	// unlock l1
	err = l1.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	err = <-c
	if err != nil {
		t.Fatal(err)
	}

	// unlock l2
	err = l2.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLockUnlockAfter(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatalf("l1 %v: %v", lockfile, err)
	}

	c := make(chan error)
	go func() {
		c <- l2.Lock(time.Second)
	}()

	time.Sleep(250 * time.Millisecond)
	// unlock l1
	err = l1.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	err = <-c
	if err != nil {
		t.Fatal(err)
	}

	// unlock l2
	err = l2.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLockTimeout(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatalf("l1 %v: %v", lockfile, err)
	}
	err = l2.Lock(time.Second)
	if err != ErrTimeout {
		t.Fatal(err)
	}

	// remove lock file
	err = l1.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLockAlreadyLocked(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	// fake out lock
	l1.descriptor = &os.File{}
	err = l1.Lock(time.Second)
	if err == nil {
		t.Fatal("lock should have failed")
	}
}

func TestUnlockAlreadyUnlocked(t *testing.T) {
	l1, err := New(lockfile)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatalf("l1 %v: %v", lockfile, err)
	}

	// fake unlock out
	fd := l1.descriptor
	l1.descriptor = nil

	err = l1.Unlock()
	if err == nil {
		t.Fatalf("unlock should have failed")
	}

	l1.descriptor = fd
	err = l1.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}
