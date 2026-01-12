# ExternalDNS Plugin CoreDNS Webhook

## Deprecated

With the release of version 0.21 or newer, every function is now included inside the in-tree provider.

## Commandline

```
usage: external-dns-coredns-webhook [<flags>]

ExternalDNS CoreDNS webhook

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --version                  Show application version.
  --dry-run                  When enabled, prints DNS record changes rather than actually performing them (default: disabled)
  --log-format=text          The format in which log messages are printed (default: text, options: text, json)
  --log-level=info           Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal
  --webhook-provider-read-timeout=5s  
                             The read timeout for the webhook provider in duration format (default: 5s)
  --webhook-provider-write-timeout=5s  
                             The write timeout for the webhook provider in duration format (default: 5s)
  --webhook-provider-port="0.0.0.0:8888"  
                             Webhook provider port (default: 0.0.0.0:8888)
  --prefix="/skydns/"        Specify the prefix name
  --managed-by=""            Only allow checking of services created by the same manager (default: "")
  --ignore-empty-managed-by  If the 'managed-by' field is set, this prevents the takeover of services without a 'managed-by' value (default: disabled)
```

## ENVs for Etcd

| Name                 | Description                                                                        | Default                 |
|----------------------|------------------------------------------------------------------------------------|-------------------------|
| ETCD_URLS            | Optionally, can be used to configure the urls to connect to etcd, comma seperated. | "http://localhost:2379" |
| ETCD_USERNAME        | Optionally, can be used to configure for authenticating to etcd.                   | ""                      | 
| ETCD_PASSWORD        | Optionally, can be used to configure for authenticating to etcd.                   | ""                      |
| ETCD_CA_FILE         | Optionally, can be used to configure TLS settings for etcd.                        | ""                      |
| ETCD_CERT_FILE       | Optionally, can be used to configure TLS settings for etcd.                        | ""                      |
| ETCD_KEY_FILE        | Optionally, can be used to configure TLS settings for etcd.                        | ""                      |
| ETCD_TLS_SERVER_NAME | Optionally, can be used to configure TLS settings for etcd.                        | ""                      |
| ETCD_TLS_INSECURE    | Optionally, To insecure handle connection use "true", default is false.            | ""                      |

## Pre-filtering CoreDNS services based on managed by field

If you are running external-dns in multi cluster, you can use `--managed-by` to filter externally created services, for example from a different external-dns.

## Custom attributes

Coredns offers currently a single custom attribute:

* [Grouped](https://github.com/skynetservices/skydns#groups)
  records: `external-dns.alpha.kubernetes.io/webhook-coredns-group`