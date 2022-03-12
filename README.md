# godns


### rule

- pattern
    - a pattern has higher priority if it has a lower index
    - `domain` with optional `record`
        - `{"domain": ["github.com"]}`
        - `{"domain": ["github.com"], "record": ["A"]}`
    - `suffix` with optional `record`
        - `{"suffix": [".com"]}`
        - `{"suffix": [".com"], "record": ["AAAA"]}`
    - https://en.wikipedia.org/wiki/List_of_DNS_record_types
- upstream
    - UDP
    - TCP
    - DoT
    - DoH
    - IPv4
    - IPv6
    - NODATA
    - NXDOMAIN
