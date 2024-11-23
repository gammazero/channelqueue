package channelqueue

import "github.com/gammazero/deque"

// ChannelQueue uses a queue to buffer data between input and output channels.
type ChannelQueue[T any] struct {
	input, output chan T
	length        chan int
	capacity      int
}

type Option[T any] func(*ChannelQueue[T])

// WithCapacity sets the limit on the number of unread items that channelqueue
// will hold. Unbuffered behavior is not supported (use a normal channel for
// that), and a value of zero or less configures the default of no limit.
//
// Example:
//
//	cq := channelqueue.New(WithCapacity[int](64))
func WithCapacity[T any](n int) func(*ChannelQueue[T]) {
	return func(c *ChannelQueue[T]) {
		if n < 1 {
			n = -1
		}
		c.capacity = n
	}
}

// WithInput uses an exist channel as the input channel.
//
// Example:
//
//	in := make(chan int)
//	cq := channelqueue.New(channlequeue.WithInput[int](in))
func WithInput[T any](in chan T) func(*ChannelQueue[T]) {
	return func(c *ChannelQueue[T]) {
		if in != nil {
			c.input = in
		}
	}
}

// New creates a new ChannelQueue that, by default, holds an unbounded number
// of items of the specified type.
func New[T any](options ...Option[T]) *ChannelQueue[T] {
	cq := &ChannelQueue[T]{
		output:   make(chan T),
		length:   make(chan int),
		capacity: -1,
	}
	for _, opt := range options {
		opt(cq)
	}
	if cq.input == nil {
		cq.input = make(chan T)
	}

	go cq.bufferData()
	return cq
}

// NewRing creates a new ChannelQueue with the specified buffer capacity, and
// circular buffer behavior. When the buffer is full, writing an additional
// item discards the oldest buffered item.
func NewRing[T any](options ...Option[T]) *ChannelQueue[T] {
	cq := &ChannelQueue[T]{
		output:   make(chan T),
		length:   make(chan int),
		capacity: -1,
	}
	for _, opt := range options {
		opt(cq)
	}
	if cq.capacity < 1 {
		// Unbounded ring is the same as an unbounded queue.
		return New(WithInput[T](cq.input))
	}
	if cq.input == nil {
		cq.input = make(chan T)
	}
	if cq.capacity == 1 {
		go cq.oneBufferData()
	} else {
		go cq.ringBufferData()
	}
	return cq
}

// In returns the write side of the channel.
func (cq *ChannelQueue[T]) In() chan<- T {
	return cq.input
}

// Out returns the read side of the channel.
func (cq *ChannelQueue[T]) Out() <-chan T {
	return cq.output
}

// Len returns the number of items buffered in the channel.
func (cq *ChannelQueue[T]) Len() int {
	return <-cq.length
}

// Cap returns the capacity of the channel.
func (cq *ChannelQueue[T]) Cap() int {
	return cq.capacity
}

// Close closes the input channel. Additional input will panic, output will
// continue to be readable until there is no more data, and then the output
// channel is closed.
func (cq *ChannelQueue[T]) Close() {
	close(cq.input)
}

// Shutdown calls Close then drains the channel to ensure that the internal
// goroutine finishes.
func (cq *ChannelQueue[T]) Shutdown() {
	close(cq.input)
	for range cq.output {
	}
}

// bufferData is the goroutine that transfers data from the In() chan to the
// buffer and from the buffer to the Out() chan.
func (cq *ChannelQueue[T]) bufferData() {
	var buffer deque.Deque[T]
	var output chan T
	var next, zero T
	inputChan := cq.input
	input := inputChan

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				// Push data from input chan to buffer.
				buffer.PushBack(elem)
			} else {
				// Input chan closed; do not select input chan.
				input = nil
				inputChan = nil
			}
		case output <- next:
			// Wrote buffered data to output chan. Remove item from buffer.
			buffer.PopFront()
		case cq.length <- buffer.Len():
		}

		if buffer.Len() == 0 {
			// No buffered data; do not select output chan.
			output = nil
			next = zero // set to zero to GC value
		} else {
			// Try to write it to output chan.
			output = cq.output
			next = buffer.Front()
		}

		if cq.capacity != -1 {
			// If buffer at capacity, then stop accepting input.
			if buffer.Len() >= cq.capacity {
				input = nil
			} else {
				input = inputChan
			}
		}
	}

	close(cq.output)
	close(cq.length)
}

// ringBufferData is the goroutine that transfers data from the In() chan to
// the buffer and from the buffer to the Out() chan, with circular buffer
// behavior of discarding the oldest item when writing to a full buffer.
func (cq *ChannelQueue[T]) ringBufferData() {
	var buffer deque.Deque[T]
	var output chan T
	var next, zero T
	input := cq.input

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				// Push data from input chan to buffer.
				buffer.PushBack(elem)
				if buffer.Len() > cq.capacity {
					buffer.PopFront()
				}
			} else {
				// Input chan closed; do not select input chan.
				input = nil
			}
		case output <- next:
			// Wrote buffered data to output chan. Remove item from buffer.
			buffer.PopFront()
		case cq.length <- buffer.Len():
		}

		if buffer.Len() == 0 {
			// No buffered data; do not select output chan.
			output = nil
			next = zero // set to zero to GC value
		} else {
			// Try to write it to output chan.
			output = cq.output
			next = buffer.Front()
		}
	}

	close(cq.output)
	close(cq.length)
}

// oneBufferData is the same as ringBufferData, but with a buffer size of 1.
func (cq *ChannelQueue[T]) oneBufferData() {
	var bufLen int
	var output chan T
	var next, zero T
	input := cq.input

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				// Push data from input chan to buffer.
				next = elem
				bufLen = 1
				// Try to write it to output chan.
				output = cq.output
			} else {
				// Input chan closed; do not select input chan.
				input = nil
			}
		case output <- next:
			// Wrote buffered data to output chan. Remove item from buffer.
			bufLen = 0
			next = zero // set to zero to GC value
			// No buffered data; do not select output chan.
			output = nil
		case cq.length <- bufLen:
		}
	}

	close(cq.output)
	close(cq.length)
}
