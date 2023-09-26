# ExternalDNS Plugin CoreDNS Webhook

## Commandline
```
usage: external-dns-coredns-webhook [<flags>]

ExternalDNS CoreDNS webhook

Flags:
  --help                    Show context-sensitive help (also try --help-long and --help-man).
  --version                 Show application version.
  --dry-run                 When enabled, prints DNS record changes rather than actually performing them (default: disabled)
  --log-format=text         The format in which log messages are printed (default: text, options: text, json)
  --log-level=info          Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal
  --webhook-provider-read-timeout=5s  
                            The read timeout for the webhook provider in duration format (default: 5s)
  --webhook-provider-write-timeout=5s  
                            The write timeout for the webhook provider in duration format (default: 5s)
  --webhook-provider-port="0.0.0.0:8888"  
                            Webhook provider port (default: 0.0.0.0:8888)
  --prefix="/skydns/"       Specify the prefix name
  --txt-owner-id="default"  When using the TXT registry, a name that identifies this instance of ExternalDNS (default: default)
  --pre-filter-external-owned-records  
                            Services are pre filter based on the txt-owner-id (default: false)
```

## Pre-filtering CoreDNS services based on ownerIDs

If you are running external-dns in multi cluster, you can use `--coredns-pre-filter-external-owned-records` and 
`--txt-owner-id` to ignore external created services, for example from a different external-dns.

## Custom attributes

Coredns offers currently a single custom attribute:

* [Grouped](https://github.com/skynetservices/skydns#groups) records: `external-dns.alpha.kubernetes.io/coredns-group`