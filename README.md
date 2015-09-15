# asyncqueue
The asyncronous queue container for GO language.

Description
-----------

The asyncqueue package implements an asynchronous FIFO queue. 
Note, standard GO channel is fast FIFO buffer. The price for a speed is no visibility on content of the buffer. This implementation allows application to see content of the queue in time consistent snapshot. Also it enqueues the calls similar to channel.

* Pull calls are enqueued in case no data.
* Push calls are enqueued in case no space.

Fare warning, this implementation is memory efficient and reasonably fast. However, it was not intended to serve massive queue with dynamic capacity or fight for a top speed under heavy race load.


Installation
------------

This package can be installed with the go get command:

    go get github.com/smile-on/asyncqueue
    
Author
------

Slav Kochepasov (a.k.a smile-on)
