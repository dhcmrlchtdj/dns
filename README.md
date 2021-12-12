# GoDNS

DNS server

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
$ cd aur && makepkg -srci
$ cat /etc/godns/config.json
$ systemctl enable --now godns.service
```

### mac
```
$ brew tap dhcmrlchtdj/godns https://github.com/dhcmrlchtdj/godns
$ brew install --HEAD dhcmrlchtdj/godns/godns
$ cat "$(brew --prefix)/etc/godns/config.json"
$ brew services start godns
```

## Config

```json
{
    "port": 1053,
    "logLevel": "info",
    "forward": [
        { "dns": "ipv4://127.0.0.1", "domain": ["localhost"] },
        { "dns": "udp://1.1.1.1:53", "domain": ["doh.pub"] },
        { "dns": "doh://1.1.1.1/dns-query", "domain": ["."], "https_proxy": "http://127.0.0.1:1080" },
        { "dns": "doh://doh.pub/dns-query", "domain": ["cn"] }
    ]
}
```

### generate accelerated-domains.china.conf

```sh
$ curl -LO 'https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf'
$ grep -v '^#' accelerated-domains.china.conf | sed -e 's|^server=/\(.*\)/114.114.114.114$|"\1",|'
```
