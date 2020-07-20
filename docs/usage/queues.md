---
layout: default
title: Queues
parent: Usage
nav_order: 7
---
# Queues

Since `0.3.5`

**this is young feature** - functionality could be dramatically enriched in a future.

By default, queues stored in a directory-based style. Each element of queue pipes directly from
 incoming requests, as well as to lambda without caching. It means - RAM usage is almost constant regardless
 of requests sizes and number of elements in a queue.

Currently, there are not security restrictions for the queue on append time. All checks will be performed before lambda
execution in same way as it defined in security. 

Queues that bound to the lambda could be found in Overview -> Endpoint page.

Queue can be re-assigned to another lambda without destroy.

In case of failure, task will be re-tried after defined interval with limited number of attempts. 0 retry means 
no **additional attempts** - at least once the task will be processed. After failure, a queue worker will wait required 
time, and it will not process other tasks.

After lambda removal, linked queues are also will be **automatically removed**.

Designed for

* to provide async processing for long-running tasks;

NOT designed for

* load balancing (but possible using multiple queues);

Endpoint: `/q/:queue-name`

Allowed queue name: latin letters, numbers and dash up. From 3 to 64 symbols length.

One queue always linked to one lambda, but one lambda can be linked to multiple queues.


