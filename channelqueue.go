package channelqueue

import "github.com/gammazero/deque"

// ChannelQueue uses a queue to buffer data between input and output channels.
type ChannelQueue[T any] struct {
	input, output chan T
	length        chan int
	capacity      int
}

// New creates a new ChannelQueue with the specified buffer capacity.
//
// A capacity < 0 specifies unlimited capacity. Unbuffered behavior is not
// supported; use a normal channel for that. Use caution if specifying an
// unlimited capacity since storage is still limited by system resources.
func New[T any](capacity int) *ChannelQueue[T] {
	if capacity == 0 {
		panic("unbuffered behavior not supported")
	}
	if capacity < 0 {
		capacity = -1
	}
	cq := &ChannelQueue[T]{
		input:    make(chan T),
		output:   make(chan T),
		length:   make(chan int),
		capacity: capacity,
	}
	go cq.bufferData()
	return cq
}

// NewRing creates a new ChannelQueue with the specified buffer capacity, and
// circular buffer behavior. When the buffer is full, writing an additional
// item discards the oldest buffered item.
func NewRing[T any](capacity int) *ChannelQueue[T] {
	if capacity < 1 {
		return New[T](capacity)
	}

	cq := &ChannelQueue[T]{
		input:    make(chan T),
		output:   make(chan T),
		length:   make(chan int),
		capacity: capacity,
	}
	if capacity == 1 {
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
