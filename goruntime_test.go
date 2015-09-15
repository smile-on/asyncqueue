// Know your runtime beofre you design the multitreading code!

package asyncqueue

import (
	"sync"
	"testing"
)

// 2.90 GHz tick = 0.345 ns i7 amd64/Win7 ; 2.0 GHz tick=0.5 Duo2 amd64/linux

// 50 ns (147t) = wholy crap -> no code structure please
func BenchmarkSimpleSubcall(b *testing.B) {
	var dummy sync.Mutex
	for i := 0; i < b.N; i++ {
		subcall(dummy)
	}
}
func subcall(e interface{}) {
	if e != nil {
	}
}

// 18 ns (52t) ; 56.3-6 ns (113t)
func BenchmarkLocking(b *testing.B) {
	var m sync.Mutex
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.Unlock()
	}
}

// ?; 347ns
func BenchmarkLockingDefered(b *testing.B) {
	n := b.N
	var m []sync.Mutex
	for i := 0; i < n; i++ {
		m = append(m, *new(sync.Mutex))
	}
	b.ResetTimer()
	for i := 0; i < n; i++ {
		m[i].Lock()
		defer m[i].Unlock()
	}
}

// 37 ns = lock *2 ;)
func BenchmarkChanWrite(b *testing.B) {
	n := b.N
	c := make(chan bool, n)
	for i := 0; i < n; i++ {
		c <- true
	}
}

// 38 ns = lock *2
func BenchmarkChanRead(b *testing.B) {
	n := b.N
	c := make(chan bool, n)
	for i := 0; i < n; i++ {
		c <- true
	}
	b.ResetTimer()
	for i := 0; i < n; i++ {
		<-c
	}
}
