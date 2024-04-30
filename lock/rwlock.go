package lock

import (
	"sync"
)

type RWLock struct {
	mutex        sync.Mutex
	cond         *sync.Cond
	activeReaders   int
	waitingWriters  int
	activeWriters   bool
}

func NewRWLock() *RWLock {
	lock := &RWLock{}
	lock.cond = sync.NewCond(&lock.mutex)
	return lock
}

func (lock *RWLock) Lock() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.waitingWriters++
	for lock.activeReaders > 0 || lock.activeWriters {
		lock.cond.Wait()
	}
	lock.waitingWriters--
	lock.activeWriters = true
}

func (lock *RWLock) Unlock() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.activeWriters = false
	lock.cond.Broadcast() // Wake up all goroutines (both readers and writers)
}

func (lock *RWLock) RLock() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	for lock.activeWriters || lock.waitingWriters > 0 || lock.activeReaders == 32 {
		lock.cond.Wait()
	}
	lock.activeReaders++
}

func (lock *RWLock) RUnlock() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.activeReaders--
	if lock.activeReaders == 0 {
		lock.cond.Broadcast() // Wake up all goroutines (both readers and writers)
	}
}