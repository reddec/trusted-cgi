---
layout: default
title: Policies
parent: Administrating
nav_order: 2
---
# Policies

Before `0.3.5` most security settings were in manifest.

Policy is a combination of restrictions described below, however, they can be applied to
several objects.

For example, you can make one policy called `my-customer-1`, that will contain several tokens,
IP restrictions etc.., and apply it to several lambdas, keeping access information in one place.

Currently, lambda can be linked to only one policy, but one policy could be linked to multiple lambdas.

## Checks

Security checks aimed to restrict access to function to the limited group of clients.

All security checks performed after application resolution. Each rule combined by AND operator.
For example, if restrictions by IP defined as well as restrictions by Origin, both limitations will
be applied for a client request.

In case of failure the 403 Access Deny will be returned.

UI:
 
1. click to any created application
2. click to security tab 

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

Restrict access by a client IP.

**Not working properly in docker container**: docker proxies all requests, so client IP will be a docker IP
instead of real address.

Should be defined in default notation `XXX.YYY.ZZZ.TTT`, IPv6 has to be supported but not tested.

The performance almost not impacted regardless number of IP. 

Since `0.3.8` the following headers are respected if `--behind-proxy` (`BEHIND_PROXY=true`) flag is set:

- `X-Real-Ip`
- `X-Forwarded-For`

The first address in chain will be used as client address.