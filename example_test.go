package asyncqueue

import (
	"fmt"
)

func Example() {
	aq := NewAsyncQueue(2)
	// queue is FIFO buffer
	aq.Push(1)
	aq.Push(2)
	if e, err := aq.Pull(); err == nil {
		fmt.Println(e)
	}
	if e, err := aq.Pull(); err == nil {
		fmt.Println(e)
	}
	// call to closed queue returns error
	aq.Close()
	if _, err := aq.Pull(); err != nil {
		fmt.Println(err)
	}

	// Output:
	// 1
	// 2
	// AsyncQueue is closed
}
