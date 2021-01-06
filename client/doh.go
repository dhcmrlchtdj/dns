package client

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

///

var dohClientPool = sync.Pool{
	New: func() interface{} {
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: 5000 * time.Millisecond,
					}
					return d.DialContext(ctx, "udp", "119.29.29.29:53")
				},
			},
		}
		customTransport := &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
		}
		httpClient := &http.Client{Transport: customTransport}
		return httpClient
	},
}

///

type DNSResponse struct {
	// The Response Code of the DNS Query.
	Status int `json:"Status"`
	// If true, it means the truncated bit was set.
	TC bool `json:"TC"`
	// If true, it means the Recursive Desired bit was set.
	RD bool `json:"RD"`
	// If true, it means the Recursion Available bit was set.
	RA bool `json:"RA"`
	// If true, it means that every record in the answer was verified with DNSSEC.
	AD bool `json:"AD"`
	// If true, the client asked to disable DNSSEC validation.
	CD       bool                  `json:"CD"`
	Question []DNSResponseQuestion `json:"Question"`
	Answer   []Answer              `json:"Answer"`
}

type DNSResponseQuestion struct {
	// The record name requested.
	Name string `json:"name"`
	// The type of DNS record requested.
	Type uint16 `json:"type"`
}

///

type DNSQuery struct {
	// Query Name
	Name string `json:"name"`
	// Query Type
	Type uint16 `json:"type"`
	// DO bit - set if client wants DNSSEC data
	DO bool `json:"do"`
	// CD bit - set to disable validation
	CD bool `json:"cd"`
}
