//todo read & write racing tests

package asyncqueue

import (
	"math/rand"
	"testing"
	"time"
)

const TIME_TAKES_TO_START_GOROUTINE = 1 * time.Millisecond // less than that

func TestErrorOnClosedQueue(t *testing.T) {
	a := NewAsyncQueue(1)
	a.Close()
	err := a.Push(1)
	if err == nil {
		t.Errorf("no expected error on push to a closed queue")
	}
	_, err = a.Pull()
	if err == nil {
		t.Errorf("no expected error on poll from a closed queue")
	}
}

func TestHoldOnPush(t *testing.T) {
	a := NewAsyncQueue(2)
	overflow := make(chan error)

	a.Push(1)
	a.Push(2)
	go func() {
		err := a.Push(3)
		overflow <- err
	}()
	select {
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		//ok
	case <-overflow:
		t.Error("overflow push has not stopped on hold")
	}
	a.Close() //release Push
	select {
	case err := <-overflow:
		if err == ERR_QUEUE_CLOSED {
			//ok
		} else {
			t.Error("push released by close has incorrect error code.")
		}
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		t.Error("close has not released waiting push")
	}
}

func TestHoldOnPull(t *testing.T) {
	overflow := make(chan error)
	a := NewAsyncQueue(2)

	go func() {
		_, err := a.Pull()
		overflow <- err
	}()
	select {
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		//ok
	case <-overflow:
		t.Error("overflow poll has not stopped on hold")
	}
	a.Close() //release Push
	select {
	case err := <-overflow:
		if err == ERR_QUEUE_CLOSED {
			//ok
		} else {
			t.Error("pull released by close has incorrect error code.")
		}
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		t.Error("close has not released waiting pull")
	}
}

func TestRun(t *testing.T) {
	a := NewAsyncQueue(2)
	if a == nil {
		t.Errorf("got nil from NewAsyncQueue()")
		t.FailNow()
	}
	defer a.Close()
	a.Push(1)
	a.Push(2)
	n, err := a.Pull()
	if err != nil {
		t.Errorf("unexpected closed queue while polling")
	} else if n != 1 {
		t.Errorf("polled val doesn't match pushed")
	}
}

func TestRunMinimalQueue(t *testing.T) {
	overflow := make(chan int)
	a := NewAsyncQueue(0)
	if a == nil {
		t.Errorf("got nil from NewAsyncQueue()")
		t.FailNow()
	}
	defer a.Close()
	//normal run
	r := rand.Float64()
	a.Push(r)
	n, err := a.Pull()
	if err != nil {
		t.Errorf("unexpected closed queue while polling")
	} else if n != r {
		t.Errorf("polled val doesn't match pushed")
	}
	//hold on push
	a.Push(1)
	go func() {
		a.Push(2)
		overflow <- 1
	}()
	select {
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		//ok
	case <-overflow:
		t.Error("overflow push has not stopped on hold")
	}
	//hold on pull
	a.Pull()   //1
	a.Pull()   //2
	<-overflow // cleanup
	go func() {
		a.Pull()
		overflow <- 1
	}()
	select {
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		//ok
	case <-overflow:
		t.Error("overflow poll has not stopped on hold")
	}
}

// 45ns with lock/unlock 18ns
func BenchmarkIdealPush(b *testing.B) {
	n := b.N
	queue := NewAsyncQueue(n + 2)
	queue.Push(nil)
	b.ResetTimer()
	for i := 0; i < n; i++ {
		queue.Push(nil)
	}
}

// 46ns + call by copy values  = 68 ns @142
// 44 + call = 102 ns @ 151
func BenchmarkNowaitPush(b *testing.B) {
	n := b.N
	queue := NewAsyncQueue(n)
	b.ResetTimer()
	for i := 0; i < n; i++ {
		queue.Push(i)
	}
}

// 45ns
func BenchmarkIdealPull(b *testing.B) {
	n := b.N
	queue := NewAsyncQueue(n + 2)
	for i := 0; i < n+2; i++ {
		queue.Push(i)
	}
	queue.Pull()
	b.ResetTimer()
	for i := 0; i < n; i++ {
		queue.Pull()
	}
}

// 42ns + type validation on return value =95 ns @142
// 42ns + validation =  53ns @151
func BenchmarkNowaitPull(b *testing.B) {
	n := b.N
	queue := NewAsyncQueue(n)
	for i := 0; i < n; i++ {
		queue.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < n; i++ {
		e, err := queue.Pull()
		if e != i {
			b.Errorf("pull value doesn't match pushed at %d with err %s", i, err)
			b.FailNow()
		}
	}
}

func TestPeak(t *testing.T) {
	a := NewAsyncQueue(2)
	defer a.Close() // 2d call on close is OK

	a.Push(1)
	if a, _ := a.Peak(); a != 1 {
		t.Error("1st peak returns wrong value.")
	}
	if a, _ := a.Peak(); a != 1 {
		t.Error("2d peak returns wrong value.")
	}
	a.Push(2)
	if a, _ := a.Peak(); a != 1 {
		t.Error("3rd peak returns wrong value.")
	}
	a.Pull()
	if a, _ := a.Peak(); a != 2 {
		t.Error("4th peak returns wrong value.")
	}
	if a, _ := a.Peak(); a != 2 {
		t.Error("5th peak returns wrong value.")
	}
	overflow := make(chan error)
	a.Pull()
	go func() {
		_, err := a.Peak()
		overflow <- err
	}()
	select {
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		//ok
	case <-overflow:
		t.Error("overflow peak has not stopped on empty queue")
	}
	a.Close() //release Peak
	select {
	case err := <-overflow:
		if err == ERR_QUEUE_CLOSED {
			//ok
		} else {
			t.Error("pull released by close has incorrect error code.")
		}
	case <-time.After(TIME_TAKES_TO_START_GOROUTINE):
		t.Error("close has not released waiting peak")
	}
}

func TestSnap(t *testing.T) {
	a := NewAsyncQueue(3)
	defer a.Close() // 2d call on close is OK

	// It does not hold on empty queue
	status := "empty queue"
	if c, _ := a.Snap(); c != nil {
		t.Error("return wrong value from", status)
	}
	// Snap returns current content of the queue with order preserved.
	a.Push(1)
	status = "queue of size 1"
	if c, err := a.Snap(); err == nil {
		if cap(c) != 1 || c[0] != 1 {
			t.Errorf("return wrong value from %s, %#v, len %d, cap %d", status, c, len(c), cap(c))
		}
	} else {
		t.Error("return err from", status, err)
	}
	// you can repeaat snap as many time as you want, tail is not moved
	if c, err := a.Snap(); err == nil {
		if cap(c) != 1 || c[0] != 1 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	a.Push(2)
	status = "queue of size 2"
	if c, err := a.Snap(); err == nil {
		if cap(c) != 2 || c[0] != 1 || c[1] != 2 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	// you can repeaat snap as many time as you want, tail is not moved
	if c, err := a.Snap(); err == nil {
		if cap(c) != 2 || c[0] != 1 || c[1] != 2 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	// Element with 0 index is first to be out of queue.
	a.Pull() // 1 Expect #2
	status = "modified queue of size 1"
	if c, err := a.Snap(); err == nil {
		if cap(c) != 1 || c[0] != 2 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	if c, err := a.Snap(); err == nil {
		if cap(c) != 1 || c[0] != 2 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	// tail ahead of head
	a.Push(3)
	a.Pull() // 2
	a.Push(4)
	status = "queue with tail ahead of head"
	if c, err := a.Snap(); err == nil {
		if cap(c) != 2 || c[0] != 3 || c[1] != 4 {
			t.Error("return wrong value from", status)
		}
	} else {
		t.Error("return err from", status, err)
	}
	// It does return error on closed queue.
	a.Close()
	status = "closed queue"
	if c, err := a.Snap(); c != nil || err != ERR_QUEUE_CLOSED {
		t.Error("return wrong value from", status)
	}
}

func TestIsEmpty(t *testing.T) {
	a := NewAsyncQueue(1)
	if !a.IsEmpty() {
		t.Errorf("new queue is not reported as empty")
	}
	a.Push(1)
	if a.IsEmpty() {
		t.Errorf("queue with one element is reported as empty")
	}
	a.Close()
	if !a.IsEmpty() {
		t.Errorf("closed queue is not reported as empty")
	}
}

func TestIsFull(t *testing.T) {
	a := NewAsyncQueue(1)
	if a.IsFull() {
		t.Errorf("new queue is reported as full")
	}
	a.Push(1)
	if !a.IsFull() {
		t.Errorf("queue with full capacity is not reported as full")
	}
	a.Close()
	if a.IsFull() {
		t.Errorf("closed queue is reported as full")
	}
}

func TestIsClosed(t *testing.T) {
	a := NewAsyncQueue(1)
	if a.IsClosed() {
		t.Errorf("new queue is reported as closed")
	}
	a.Push(1)
	if a.IsClosed() {
		t.Errorf("queue with one element is reported as closed")
	}
	a.Close()
	if !a.IsClosed() {
		t.Errorf("closed queue is not reported as closed")
	}
}
