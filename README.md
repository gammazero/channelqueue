# channelqueue
[![GoDoc](https://pkg.go.dev/badge/github.com/gammazero/channelqueue)](https://pkg.go.dev/github.com/gammazero/channelqueue)
[![Build Status](https://github.com/gammazero/channelqueue/actions/workflows/go.yml/badge.svg)](https://github.com/gammazero/channelqueue/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gammazero/channelqueue)](https://goreportcard.com/report/github.com/gammazero/channelqueue)
[![codecov](https://codecov.io/gh/gammazero/channelqueue/branch/master/graph/badge.svg)](https://codecov.io/gh/gammazero/channelqueue)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Concurrently access a dynamic queue using channels.

ChannelQueue implements a queue that uses channels for input and output to provide concurrent access to a dynamically-sized queue. This allows the queue to be used like a channel, in a thread-safe manner. Closing the input channel closes the output channel when all queued items are read, consistent with channel behavior. In other words a ChannelQueue is a dynamically buffered channel with up to infinite capacity.

ChannelQueue also supports circular buffer behavior when created using `NewRing`. When the buffer is full, writing an additional item discards the oldest buffered item.

When specifying an unlimited buffer capacity use caution as the buffer is still limited by the resources available on the host system.

The ChannelQueue buffer auto-resizes according to the number of items buffered. For more information on the internal queue, see: https://github.com/gammazero/deque

ChannelQueue uses generics to contain items of the type specified. To create a ChannelQueue that holds a specific type, provide a type argument to `New`. For example:
```go
intChanQueue := channelqueue.New[int]()
stringChanQueue := channelqueue.New[string]()
```

ChannelQueue can be used to provide buffering between existing channels. Using an existing read chan, write chan, or both are supported:
```go
in := make(chan int)
out := make(chan int)

// Create a buffer between in and out channels.
channelqueue.New(channelqueue.WithInput[int](in), channelqueue.WithOutput[int](out))
// ...
close(in) // this will close cq when all output is read.
```
