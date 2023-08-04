package channelqueue_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	cq "github.com/gammazero/channelqueue"
)

func TestCapLen(t *testing.T) {
	ch := cq.New[int](-1)
	if ch.Cap() != -1 {
		t.Error("expected capacity -1")
	}

	ch = cq.New[int](3)
	if ch.Cap() != 3 {
		t.Error("expected capacity 3")
	}

	if ch.Len() != 0 {
		t.Error("expected 0 from Len()")
	}
	in := ch.In()
	for i := 0; i < ch.Cap(); i++ {
		if ch.Len() != i {
			t.Errorf("expected %d from Len()", i)
		}
		in <- i
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from capacity 0")
		}
	}()
	ch = cq.New[int](0)
	if ch != nil {
		t.Fatal("expected nil")
	}
}

func TestUnlimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := cq.New[int](-1)
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
	ch := cq.New[int](32)
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
	ch := cq.New[int](32)
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
	ch := cq.New[int](-1)

	var err error
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			if ch.Len() > 1000 {
				err = errors.New("Len too great")
			}
			if ch.Cap() != -1 {
				err = errors.New("expected Cap to return -1")
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
	if err != nil {
		t.Fatal(err)
	}
}

func TestDouble(t *testing.T) {
	const msgCount = 1000
	ch := cq.New[int](100)
	recvCh := cq.New[int](100)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	var err error
	go func() {
		for i := 0; i < msgCount; i++ {
			val := <-ch.Out()
			if i != val {
				err = fmt.Errorf("expected %d but got %d", i, val)
				return
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
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeadlock(t *testing.T) {
	ch := cq.New[int](1)
	ch.In() <- 1
	<-ch.Out()

	done := make(chan struct{})
	go func() {
		ch.In() <- 2
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Fatal("could not write to channel")
	}
}

func TestRing(t *testing.T) {
	ch := cq.NewRing[rune](5)
	for _, r := range "hello" {
		ch.In() <- r
	}

	ch.In() <- 'w'
	char := <-ch.Out()
	if char != 'e' {
		t.Fatal("expected 'e' but got", char)
	}

	for _, r := range "abcdefghij" {
		ch.In() <- r
	}

	ch.Close()

	out := make([]rune, 0, ch.Len())
	for r := range ch.Out() {
		out = append(out, r)
	}
	if string(out) != "fghij" {
		t.Fatalf("expected \"fghij\" but got %q", out)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from capacity 0")
		}
	}()
	ch = cq.NewRing[rune](0)
	if ch != nil {
		t.Fatal("expected nil")
	}
}

func TestRingOne(t *testing.T) {
	ch := cq.NewRing[rune](1)
	for _, r := range "hello" {
		ch.In() <- r
	}

	ch.In() <- 'w'
	if ch.Len() != 1 {
		t.Fatalf("expected length 1, got %d", ch.Len())
	}
	char := <-ch.Out()
	if char != 'w' {
		t.Fatal("expected 'w' but got", char)
	}
	if ch.Len() != 0 {
		t.Fatal("expected length 0")
	}

	for _, r := range "abcdefghij" {
		ch.In() <- r
	}

	ch.Close()

	out := make([]rune, 0, ch.Len())
	for r := range ch.Out() {
		out = append(out, r)
	}
	if string(out) != "j" {
		t.Fatalf("expected \"j\" but got %q", out)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from capacity 0")
		}
	}()
	ch = cq.NewRing[rune](0)
	if ch != nil {
		t.Fatal("expected nil")
	}
}

func BenchmarkSerial(b *testing.B) {
	ch := cq.New[int](b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- i
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkParallel(b *testing.B) {
	ch := cq.New[int](b.N)
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
	ch := cq.New[int](b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- i
		<-ch.Out()
	}
}
