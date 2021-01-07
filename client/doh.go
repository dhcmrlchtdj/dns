package client

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

///

var httpClient = func() *http.Client {
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
	return &http.Client{Transport: customTransport}
}()

func GenDoHClient(dohServer string) DNSClient {
	// https://cloudflare-dns.com/dns-query
	// https://doh.pub/dns-query
	return func(name string, qtype uint16) []Answer {
		req, err := http.NewRequest("GET", dohServer, nil)
		if err != nil {
			log.Println(err)
			return nil
		}
		req.Header.Set("accept", "application/dns-json")
		q := req.URL.Query()
		q.Set("name", name)                     // Query Name
		q.Set("type", dns.Type(qtype).String()) // Query Type
		q.Set("do", "false")                    // DO bit - set if client wants DNSSEC data
		q.Set("cd", "false")                    // CD bit - set to disable validation
		req.URL.RawQuery = q.Encode()

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Println(err)
			return nil
		}
		defer resp.Body.Close()

		var r dohResponse
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			log.Println(err)
			return nil
		}

		// TODO check r.Status

		return r.Answer
	}
}

///

type dohResponse struct {
	Status   int  `json:"Status"` // The Response Code of the DNS Query.
	TC       bool `json:"TC"`     // If true, it means the truncated bit was set.
	RD       bool `json:"RD"`     // If true, it means the Recursive Desired bit was set.
	RA       bool `json:"RA"`     // If true, it means the Recursion Available bit was set.
	AD       bool `json:"AD"`     // If true, it means that every record in the answer was verified with DNSSEC.
	CD       bool `json:"CD"`     // If true, the client asked to disable DNSSEC validation.
	Question []struct {
		Name string `json:"name"` // The record name requested.
		Type uint16 `json:"type"` // The type of DNS record requested.
	} `json:"Question"`
	Answer []Answer `json:"Answer"`
}
