# GoDNS

DNS with china list

## Example

```
$ go build
$ ./godns --help
$ ./godns --port=1053 --log-level=debug --conf=/path/to/config
```

## Usage

### arch
```
$ git clone
$ cd aur && makepkg -src
```

### mac
```
$ brew install --HEAD dhcmrlchtdj/godns/godns
$ brew services start godns
```

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

```sh
$ curl -L 'https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf' \
    -o accelerated-domains.china.conf
$ cat accelerated-domains.china.conf \
    | sed -e 's|^server=/\(.*\)/114.114.114.114$|\1|' \
    | egrep -v '^#' \
    | sed -e 's|\(.*\)|"\1",|' \
    > dns.conf
```
