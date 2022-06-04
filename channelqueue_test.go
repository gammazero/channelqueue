package channelqueue

import (
	"testing"
	"time"
)

func TestCapLen(t *testing.T) {
	cq := New[int](-1)
	if cq.Cap() != -1 {
		t.Error("expected capacity -1")
	}

	cq = New[int](3)
	if cq.Cap() != 3 {
		t.Error("expected capacity 3")
	}

	if cq.Len() != 0 {
		t.Error("expected 0 from Len()")
	}
	in := cq.In()
	for i := 0; i < cq.Cap(); i++ {
		if cq.Len() != i {
			t.Errorf("expected %d from Len()", i)
		}
		in <- i
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from capacity 0")
		}
	}()
	cq = New[int](0)
}

func TestUnlimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New[int](-1)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val {
			t.Fatal("expected", i, "but got", val)
		}
	}
}

func TestLimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New[int](32)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val {
			t.Fatal("expected", i, "but got", val)
		}
	}
}

func TestBufferLimit(t *testing.T) {
	ch := New[int](32)
	for i := 0; i < ch.Cap(); i++ {
		ch.In() <- i
	}
	var timeout bool
	select {
	case ch.In() <- 999:
	case <-time.After(200 * time.Millisecond):
		timeout = true
	}
	if !timeout {
		t.Fatal("expected timeout on full channel")
	}
}

func TestRace(t *testing.T) {
	ch := New[int](-1)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			if ch.Len() > 1000 {
				t.Fatal("Len too great")
			}
			if ch.Cap() != -1 {
				t.Fatal("expected Cap to return -1")
			}
		}
	}()

	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	go func() {
		ready <- struct{}{}
		<-start
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
	}()

	var val int
	go func() {
		ready <- struct{}{}
		<-start
		for i := 0; i < 1000; i++ {
			val = <-ch.Out()
		}
		close(done)
	}()

	<-ready
	<-ready
	close(start)
	<-done
	if val != 999 {
		t.Fatalf("last value should be 999, got %d", val)
	}
}

func TestDouble(t *testing.T) {
	const msgCount = 1000
	ch := New[int](100)
	recvCh := New[int](100)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	go func() {
		for i := 0; i < msgCount; i++ {
			val := <-ch.Out()
			if i != val {
				t.Fatal("expected", i, "but got", val)
			}
			recvCh.In() <- i
		}
	}()
	for i := 0; i < msgCount; i++ {
		val := <-recvCh.Out()
		if i != val {
			t.Fatal("expected", i, "but got", val)
		}
	}
}

func BenchmarkSerial(b *testing.B) {
	ch := New[int](b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkParallel(b *testing.B) {
	ch := New[int](b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch.Out()
		}
		<-ch.Out()
	}()
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	ch.Close()
}

func BenchmarkPushPull(b *testing.B) {
	ch := New[int](b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- i
		<-ch.Out()
	}
}
