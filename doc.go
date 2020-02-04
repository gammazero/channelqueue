/*
Package channelqueue implements a queue that uses channels for input and output
to provide concurrent access to a resizable queue, and allowing the queue to be
used like a channel. Closing the input channel closes the output channel when
all queued items are read, consistent with channel behavior.  In other words
channelqueue is a dynamically buffered channel with up to infinite capacity.

When specifying an unlimited buffer capacity use caution as the buffer is still
limited by the resources available on the host system.

Caution

The behavior of channelqueue differs from the behavior of a normal channel in
one important way: After writing to the In() channel, the data may not be
immediately available on the Out() channel (until the buffer goroutine is
scheduled), and may be missed by a non-blocking select.

Credits

This implementation is based on ideas/examples from:
https://github.com/eapache/channels

*/
package channelqueue
