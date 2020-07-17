# Queues

Since `0.3.5`

**this is young feature** - functionality could be dramatically enriched in a future.

By default, queues stored in a directory-based style. Each element of queue pipes directly from
 incoming requests, as well as to lambda without caching. It means - RAM usage is almost constant regardless
 of requests sizes and number of elements in a queue.
 

Designed for

* to provide async processing for long-running tasks;

NOT designed for

* load balancing (but possible using multiple queues);
* failure-tolerance: failed request will not be re-queued;


Endpoint: `/q/:queue-name`

Allowed queue name: latin letters, numbers and dash up. From 3 to 64 symbols length.

One queue always linked to one lambda, but one lambda can be linked to multiple queues.

