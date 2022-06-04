# channelqueue
[![GoDoc](https://pkg.go.dev/badge/github.com/gammazero/channelqueue)](https://pkg.go.dev/github.com/gammazero/channelqueue)
[![Go Report Card](https://goreportcard.com/badge/github.com/gammazero/channelqueue)](https://goreportcard.com/report/github.com/gammazero/channelqueue)
[![codecov](https://codecov.io/gh/gammazero/channelqueue/branch/master/graph/badge.svg)](https://codecov.io/gh/gammazero/channelqueue)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Concurrently access a dynamic queue using channels.

ChannelQueue implements a queue that uses channels for input and output to provide concurrent access to a dynamically-sized queue. This allows the queue to be used like a channel, in a thread-safe manner. Closing the input channel closes the output channel when all queued items are read, consistent with channel behavior. In other words a ChannelQueue is a dynamically buffered channel with up to infinite capacity.

When specifying an unlimited buffer capacity use caution as the buffer is still limited by the resources available on the host system.

The ChannelQueue buffer is supplied by a fast queue implementation, which auto-resizes according to the number of items buffered. For more information on the queue, see: https://github.com/gammazero/deque

ChannelQueue uses generics to contain items of the type specified. To create a ChannelQueue that holds a specific type, provide a type argument to `New`. For example:
```go
	intChanQueue := channelqueue.New[int](1024)
    stringChanQueue := channelqueue.New[string](-1)
```
