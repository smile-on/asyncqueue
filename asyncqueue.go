// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package asyncqueue implements an asynchronous FIFO queue.
// Pull calls are enqueued in case no data.
// Push calls are enqueued in case no space.
package asyncqueue

import (
	"fmt"
	"sync"
)

var (
	ERR_QUEUE_CLOSED = fmt.Errorf("AsyncQueue is closed")
)

const nowhere = -1

// AsyncQueue is thread safe FIFO queue with fixed capacity.
type AsyncQueue struct {
	read, write sync.Mutex
	content     *sync.Cond // for enqueed calls: readEmpty, writeFull
	access      sync.Mutex
	closed      bool
	buff        []interface{}
	head        int
	tail        int
	maxIdx      int
}

// Push inserts element to the tail of the queue
// blocks until the space is available.
// Returns error on closed queue.
func (a *AsyncQueue) Push(e interface{}) error {
	a.write.Lock()
	a.access.Lock()
	if a.closed {
		a.access.Unlock()
		a.write.Unlock()
		return ERR_QUEUE_CLOSED
	}
	if a.head == a.tail { //writeFull
		a.content.Wait()
		if a.closed { // waken by close()
			a.access.Unlock()
			a.write.Unlock()
			return ERR_QUEUE_CLOSED
		}
	}
	// write & move head
	a.buff[a.head] = e
	if a.tail == nowhere { // not empty anymore
		a.tail = a.head
		a.content.Signal()
	}
	if a.head == a.maxIdx {
		a.head = 0
	} else {
		a.head++
	}
	a.access.Unlock()
	a.write.Unlock()
	return nil
}

// Pull returns next element and removes it from queue
// blocks until the data is available.
// Returns error on closed queue.
func (a *AsyncQueue) Pull() (e interface{}, err error) {
	a.read.Lock()
	a.access.Lock()
	if a.closed {
		a.access.Unlock()
		a.read.Unlock()
		return nil, ERR_QUEUE_CLOSED
	}
	if a.tail == nowhere { //readEmpty
		a.content.Wait()
		if a.closed { // waken by close()
			a.access.Unlock()
			a.read.Unlock()
			return nil, ERR_QUEUE_CLOSED
		}
	}
	//read & move tail
	e = a.buff[a.tail]
	if a.tail == a.head { // have made first available slot
		a.content.Signal()
	}
	if a.tail == a.maxIdx {
		a.tail = 0
	} else {
		a.tail++
	}
	if a.tail == a.head { // have read last one
		a.tail = nowhere
	}
	a.access.Unlock()
	a.read.Unlock()
	return
}

// Peak is similar to Pull but does not remove element from the queue.
func (a *AsyncQueue) Peak() (e interface{}, err error) {
	a.read.Lock()
	a.access.Lock()
	if a.closed {
		a.access.Unlock()
		a.read.Unlock()
		return nil, ERR_QUEUE_CLOSED
	}
	if a.tail == nowhere { //readEmpty
		a.content.Wait()
		if a.closed { // waken by close()
			a.access.Unlock()
			a.read.Unlock()
			return nil, ERR_QUEUE_CLOSED
		}
	}
	//read & NOT move tail
	e = a.buff[a.tail]
	a.access.Unlock()
	a.read.Unlock()
	return
}

// NewAsyncQueue created queue of size n elements.
// n must be >= 1
func NewAsyncQueue(n int) *AsyncQueue {
	if n < 1 {
		n = 1
	}
	a := new(AsyncQueue)
	a.buff = make([]interface{}, n)
	a.maxIdx = n - 1
	a.tail = nowhere
	a.content = sync.NewCond(&a.access)
	return a
}

// Close realeases all application threads waiting on push or poll.
func (a *AsyncQueue) Close() {
	a.access.Lock()
	a.closed = true
	a.content.Signal() //wake the waiting threads to exit
	a.access.Unlock()
	return
}

// Snap returns a copy of current content of the queue with order preserved.
// Element with 0 index is first to be out of queue.
// It does not hold on empty queue but does return error on closed queue.
func (a *AsyncQueue) Snap() (c []interface{}, err error) {
	a.access.Lock()
	if a.closed {
		a.access.Unlock()
		return nil, ERR_QUEUE_CLOSED
	}
	if a.tail == nowhere { //readEmpty
		a.access.Unlock()
		return nil, nil
	}
	//make a copy
	size := a.maxIdx + 1
	switch len := a.head - a.tail; {
	case len > 0:
		//normal
		c = make([]interface{}, len)
		copy(c, a.buff[a.tail:a.head])
	case len < 0:
		// tail ahead of head
		len += size
		c = make([]interface{}, len)
		nc := copy(c, a.buff[a.tail:size])
		copy(c[nc:], a.buff[0:a.head])
	case len == 0:
		//buf is full
		c = make([]interface{}, size)
		copy(c, a.buff)
	}
	a.access.Unlock()
	return
}

// IsClosed return true only if queue been closed.
func (a *AsyncQueue) IsClosed() bool {
	a.access.Lock()
	is := a.closed
	a.access.Unlock()
	return is
}

// IsEmpty returns true if no data elements are in the queue.
// Closed queue is shown as empty, you can't read it.
func (a *AsyncQueue) IsEmpty() bool {
	a.access.Lock()
	is := true
	if !a.closed {
		is = (a.tail == nowhere)
	}
	a.access.Unlock()
	return is
}

// IsFull returns true if maximum capacity is reached.
// Closed queue is not full, you can't read it.
func (a *AsyncQueue) IsFull() bool {
	a.access.Lock()
	is := false
	if !a.closed {
		is = (a.tail == a.head)
	}
	a.access.Unlock()
	return is
}
