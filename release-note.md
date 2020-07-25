It's a big release, that moving the project forward to the first major release.

More stability, more security, more features, but keeping system resources usage less than ever.

### Queues

Make a simple POST request to the endpoint and let the platform manage requests asynchronously. [See docs](https://trusted-cgi.reddec.net/usage/queues/).

It doesn't matter how many messages will be enqueued - it will almost not affect memory (RAM), because
all items offloaded to the permanent storage (HDD/SSD/...). 

### Policies

No more sensitive information in a manifest - all security-related parameters now moved to platform level.
As a bonus - different lambdas now can use the same security rules (policies). [See docs](https://trusted-cgi.reddec.net/administrating/policies/)

### UI

UI refactored to provide more clean navigation for instances with a large number of objects.


## Migration notices

All manifests should migrate automatically after the restart, however, backup is always a good idea.  