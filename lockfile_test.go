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
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l1.Lock(time.Second)
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
			err := l2.Lock(time.Second)
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

func TestRaceVariable(t *testing.T) {
	l1, e1 := New(lockfile, 100*time.Millisecond)
	if e1 != nil {
		t.Fatal(e1)
	}
	l2, e2 := New(lockfile, 100*time.Millisecond)
	if e2 != nil {
		t.Fatal(e2)
	}

	var wg sync.WaitGroup
	x := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l1.Lock(5 * time.Second)
			if err != nil {
				t.Fatalf("l1 %v: %v", lockfile, err)
			}
			x++
			err = l1.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l2.Lock(5 * time.Second)
			if err != nil {
				t.Fatalf("l2 %v: %v", lockfile, err)
			}
			x--
			err = l2.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
	if x != 0 {
		t.Fatalf("invalid x %v", x)
	}
}

func TestLockUnlockRace(t *testing.T) {
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile, 100*time.Millisecond)
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
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile, 100*time.Millisecond)
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
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile, 100*time.Millisecond)
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

func TestLockExclusion(t *testing.T) {
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = l1.Lock(time.Second)
		if err != nil {
			t.Fatalf("first lock %v: %v", lockfile, err)
		}
		time.Sleep(time.Second)
		err = l1.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = l1.Lock(2 * time.Second)
		if err != nil {
			t.Fatalf("second lock %v: %v", lockfile, err)
		}
		time.Sleep(time.Second)
		err = l1.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}()
	wg.Wait()
}

func TestLockExclusionTimeout(t *testing.T) {
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond)
		err := l1.Lock(time.Second)
		if err == nil {
			t.Fatalf("first lock should have timed out")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := l1.Lock(1 * time.Second)
		if err != nil {
			t.Fatalf("second lock %v: %v", lockfile, err)
		}
		time.Sleep(2 * time.Second)
		err = l1.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}()
	wg.Wait()
}

func TestLockTimeout(t *testing.T) {
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := New(lockfile, 100*time.Millisecond)
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
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Unlock() // clean up after ourselves

	// lock against self
	err = l1.Lock(time.Second)
	if err == nil {
		t.Fatal("lock should have failed")
	}
}

func TestUnlockAlreadyUnlocked(t *testing.T) {
	l1, err := New(lockfile, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Lock(time.Second)
	if err != nil {
		t.Fatalf("l1 %v: %v", lockfile, err)
	}

	err = l1.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Unlock()
	if err == nil {
		t.Fatalf("unlock should have failed")
	}
}
