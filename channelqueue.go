package channelqueue

import "github.com/gammazero/deque"

// ChannelQueue uses a queue to buffer data between input and output channels.
type ChannelQueue struct {
	input, output chan interface{}
	length        chan int
	buffer        deque.Deque
	capacity      int
}

// New creates a new ChannelQueue with the specified buffer capacity.
//
// A capacity < 0 specifies unlimited capacity.  Unbuffered behavior is not
// supported; use a normal channel for that.  Use caution if specifying an
// unlimited capacity since storage is still limited by system resources.
func New(capacity int) *ChannelQueue {
	if capacity == 0 {
		panic("unbuffered behavior not supported")
	}
	if capacity < 0 {
		capacity = -1
	}
	cq := &ChannelQueue{
		input:    make(chan interface{}),
		output:   make(chan interface{}),
		length:   make(chan int),
		capacity: capacity,
	}
	go cq.bufferInput()
	return cq
}

// In returns the write side of the channel.
func (cq *ChannelQueue) In() chan<- interface{} {
	return cq.input
}

// Out returns the read side of the channel.
func (cq *ChannelQueue) Out() <-chan interface{} {
	return cq.output
}

// Len returns the number of items buffered in the channel.
func (cq *ChannelQueue) Len() int {
	return <-cq.length
}

// Cap returns the capacity of the channel.
func (cq *ChannelQueue) Cap() int {
	return cq.capacity
}

// Close closes the channel.  Additional input will panic, output will continue
// to be readable until nil.
func (cq *ChannelQueue) Close() {
	close(cq.input)
}

// bufferInput is the goroutine that transfers data from the In() chan to the
// buffer and from the buffer to the Out() chan.
func (cq *ChannelQueue) bufferInput() {
	var input, output, inputChan chan interface{}
	var next interface{}
	inputChan = cq.input
	input = inputChan

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				// Push data from input chan to buffer.
				cq.buffer.PushBack(elem)
			} else {
				// Input chan closed; do not select input chan.
				input = nil
				inputChan = nil
			}
		case output <- next:
			// Wrote buffered data to output chan.  Remove item from buffer.
			cq.buffer.PopFront()
		case cq.length <- cq.buffer.Len():
		}

		if cq.buffer.Len() == 0 {
			// No buffered data; do not select output chan.
			output = nil
			next = nil
		} else {
			// Try to write it to output chan.
			output = cq.output
			next = cq.buffer.Front()
			if cq.capacity != -1 {
				// If buffer at capacity, then stop accepting input.
				if cq.buffer.Len() >= cq.capacity {
					input = nil
				} else {
					input = inputChan
				}
			}
		}
	}

	close(cq.output)
	close(cq.length)
}
