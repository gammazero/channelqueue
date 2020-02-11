# channelqueue
[![Build Status](https://travis-ci.com/gammazero/channelqueue.svg)](https://travis-ci.com/gammazero/channelqueue)
[![Go Report Card](https://goreportcard.com/badge/github.com/gammazero/channelqueue)](https://goreportcard.com/report/github.com/gammazero/channelqueue)
[![codecov](https://codecov.io/gh/gammazero/channelqueue/branch/master/graph/badge.svg)](https://codecov.io/gh/gammazero/channelqueue)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Concurrently access dynamic queue using channels.

[![GoDoc](https://godoc.org/github.com/gammazero/channelqueue?status.svg)](https://godoc.org/github.com/gammazero/channelqueue)

channelqueue implements a queue that uses channels for input and output to provide concurrent access to a dynamically-sized queue, and allowing the queue to be used like a channel. Closing the input channel closes the output channel when all queued items are read, consistent with channel behavior.  In other words channelqueue is a dynamically buffered channel with up to infinite capacity.

When specifying an unlimited buffer capacity use caution as the buffer is still limited by the resources available on the host system.

The channelqueue buffer is supplied by a fast queue implementation, which auto-resizes according to the number of items buffered. For more information on the queue, see: https://github.com/gammazero/deque


