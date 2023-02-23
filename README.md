# ExternalDNS Plugin CoreDNS Provider


## Pre-filtering CoreDNS services based on ownerIDs

If your are running external-dns in multi cluster, you can use `--coredns-pre-filter-external-owned-records` and `--txt-owner-id` to ignore external created services, for example from a different external-dns.

## Custom attributes

Coredns offers currently a single custom attribute:

* [Grouped](https://github.com/skynetservices/skydns#groups) records: `external-dns.alpha.kubernetes.io/coredns-group`