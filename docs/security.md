# Security

All security checks performed after application resolution. Each rule combined by AND operator.
For example, if restrictions by IP defined as well as restrictions by Origin, both limitations will
be applied for a client request.

In case of failure the 403 Access Deny will be returned.

### Tokens

Restrict incoming requests by `Authorization` header.

Header should contain one of defined tokens.

If `public` flag is true, the setting will be ignored. 

By default, UI generates tokens by UUIDv4 algorithm, however arbitrary text could be used and defined during setup.

The performance almost not impacted regardless number of tokens. 

### Origin

Restrict access by `Origin` header. Useful to from where (domains) browser clients could access
function. 

If at least one origin defined, the security check becomes mandatory.

By standard, the field should contain a domain with protocol (ex: https://example.com) however backend is
not checking validity of the content and arbitrary text could be used and defined during setup.

Wildcards not supported.

The performance almost not impacted regardless number of origins. 

### IP

Restrict access by client IP.

**Not working properly in docker container**: docker proxies all requests, so client IP will be a docker IP
instead of real address.

Should be defined in default notation `XXX.YYY.ZZZ.TTT`, IPv6 has to be supported but not tested.

The performance almost not impacted regardless number of IP. 

## Planned features:

* 'security group' entity that aggregates security settings and could be linked to
many applications (one-to-many).