package queue

import (
	"sync/atomic"
)

type Request struct {
	Command string
	ID int 
	Body string 
	Timestamp float64
} 

type node struct {
	value *Request
	next  atomic.Value 
}

type LockFreeQueue struct {
	head atomic.Value 
	tail atomic.Value 
}

func NewLockFreeQueue() *LockFreeQueue {
	n := &node{}
	q := &LockFreeQueue{}
	q.head.Store(n)
	q.tail.Store(n)
	return q
}

func (q *LockFreeQueue) Enqueue(task *Request) {
	n := &node{value: task}
	for {
		tail := q.tail.Load().(*node)
		next := tail.next.Load()
		if tail == q.tail.Load().(*node) { // Ensure tail hasn't moved
			if next == nil {
				if tail.next.CompareAndSwap(next, n) {
					q.tail.CompareAndSwap(tail, n) // Move tail
					return
				}
			} else {
				q.tail.CompareAndSwap(tail, next.(*node)) 
			}
		}
	}
}

func (q *LockFreeQueue) Dequeue() *Request {
	for {
		head := q.head.Load().(*node)
		tail := q.tail.Load().(*node)
		next := head.next.Load()
		if head == q.head.Load().(*node) { // Ensure head hasn't moved
			if head == tail {
				if next == nil {
					return nil 
				}
				q.tail.CompareAndSwap(tail, next.(*node)) // Tail is behind, move it forward
			} else {
				val := next.(*node).value
				if q.head.CompareAndSwap(head, next.(*node)) {
					return val
				}
			}
		}
	}
}