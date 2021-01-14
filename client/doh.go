package client

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/miekg/dns"
)

///

var (
	dohHttpClient  = new(http.Client)
	dohClientCache = new(sync.Map)
)

func GetDoHClient(dohServer string) dnsClient {
	c, found := dohClientCache.Load(dohServer)
	if found {
		return c.(dnsClient)
	}

	cc := func(name string, qtype uint16) []Answer {
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

		resp, err := dohHttpClient.Do(req)
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
	dohClientCache.Store(dohServer, cc)
	return cc
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
