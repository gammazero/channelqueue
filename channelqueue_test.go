package channelqueue

import (
	"testing"
	"time"
)

func TestCapLen(t *testing.T) {
	cq := New(-1)
	if cq.Cap() != -1 {
		t.Error("expected capacity -1")
	}

	cq = New(3)
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
	cq = New(0)
}

func TestUnlimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New(-1)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func TestLimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New(32)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func TestBufferLimit(t *testing.T) {
	ch := New(32)
	for i := 0; i < ch.Cap(); i++ {
		ch.In() <- nil
	}
	var timeout bool
	select {
	case ch.In() <- nil:
	case <-time.After(200 * time.Millisecond):
		timeout = true
	}
	if !timeout {
		t.Fatal("expected timeout on full channel")
	}
}

func TestRace(t *testing.T) {
	ch := New(-1)
	go ch.Len()
	go ch.Cap()

	go func() {
		ch.In() <- nil
	}()

	go func() {
		<-ch.Out()
	}()
}

func TestDouble(t *testing.T) {
	const msgCount = 1000
	ch := New(100)
	recvCh := New(100)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	go func() {
		for i := 0; i < msgCount; i++ {
			val := <-ch.Out()
			if i != val.(int) {
				t.Fatal("expected", i, "but got", val.(int))
			}
			recvCh.In() <- i
		}
	}()
	for i := 0; i < msgCount; i++ {
		val := <-recvCh.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func BenchmarkSerial(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkParallel(b *testing.B) {
	ch := New(b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch.Out()
		}
		<-ch.Out()
	}()
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	ch.Close()
}

func BenchmarkPushPull(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		<-ch.Out()
	}
}
