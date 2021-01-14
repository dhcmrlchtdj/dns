# shunt

## Config

```json
{
    "server": [
        { "dns": "file:///etc/hosts" },
        { "dns": "ip://127.0.0.1", "domain": ["localhost"] },
        { "dns": "udp://1.1.1.1:53", "domain": ["cloudflare-dns.com"] },
        { "dns": "tcp://1.1.1.1:53", "domain": ["cloudflare-dns.com"] },
        { "dns": "dot://1.1.1.1:853", "domain": ["cloudflare-dns.com"] },
        { "dns": "doh://cloudflare-dns.com/dns-query", "domain": ["."] },
        { "dns": "doh://doh.pub/dns-query", "domain": ["cn"] }
    ]
}
```
