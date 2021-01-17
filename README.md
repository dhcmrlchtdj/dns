# shunt

## Config

```json
{
    "port": 1053,
    "logLevel": "info", // trace,debug,info,error
    "server": [
        { "dns": "file:///etc/hosts" },
        { "dns": "ipv4://127.0.0.1", "domain": ["localhost"] },
        { "dns": "udp://1.1.1.1:53", "domain": ["cloudflare-dns.com", "doh.pub"] },
        { "dns": "tcp://1.1.1.1:53", "domain": [] },
        { "dns": "dot://1.1.1.1:853", "domain": [] },
        { "dns": "doh://cloudflare-dns.com/dns-query", "domain": ["."] },
        { "dns": "doh://doh.pub/dns-query", "domain": ["cn"] }
    ]
}
```
