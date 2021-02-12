# shunt

DNS forwarder.

## Example

```
$ go build
$ ./shunt --help
$ ./shunt --port=1053 --log-level=trace --conf=/path/to/config
```

## Install

- `brew install --HEAD dhcmrlchtdj/custom-tap/shunt`

## Config

```json
{
    "port": 1053,
    "logLevel": "info",
    "forward": [
        { "dns": "ipv4://127.0.0.1", "domain": ["localhost"] },
        { "dns": "udp://1.1.1.1:53", "domain": ["cloudflare-dns.com", "doh.pub"] },
        { "dns": "doh://cloudflare-dns.com/dns-query", "domain": ["."] },
        { "dns": "doh://doh.pub/dns-query", "domain": ["cn"] }
    ]
}
```

### generate domain list

```
$ curl -L 'https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf' \
    -o accelerated-domains.china.conf
$ gsed -e 's|^server=/\(.*\)/114.114.114.114$$|\1|' \
    accelerated-domains.china.conf \
    | egrep -v '^#' \
    > accelerated-domains.china.raw.txt
$ gsed -e 's|\(.*\)|"\1",|' \
    accelerated-domains.china.raw.txt \
    > shunt.conf
```
