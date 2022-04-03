package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

var dohClientCache = new(sync.Map)

func GetDoHClient(dohServer string, proxy string) dnsClient {
	serverKey := dohServer + "-" + proxy
	c, found := dohClientCache.Load(serverKey)
	if found {
		return c.(dnsClient)
	}

	dohHttpClient := new(http.Client)
	if len(proxy) > 0 {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			panic(err)
		}
		dohHttpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	cc := func(name string, qtype uint16) []dns.RR {
		sublogger := log.With().
			Str("module", "client.doh").
			Str("server", dohServer).
			Str("proxy", proxy).
			Str("domain", name).
			Uint16("type", qtype).
			Logger()

		sublogger.Info().Msg("query")

		req, err := http.NewRequest("GET", dohServer, http.NoBody)
		if err != nil {
			sublogger.Error().Err(err).Send()
			return nil
		}
		req.Header.Set("accept", "application/dns-json")
		q := req.URL.Query()
		q.Set("name", name)                     // Query Name
		q.Set("type", dns.Type(qtype).String()) // Query Type
		// q.Set("do", "true")                     // DO bit - set if client wants DNSSEC data
		// q.Set("cd", "true")                     // CD bit - set to disable validation
		req.URL.RawQuery = q.Encode()

		resp, err := dohHttpClient.Do(req)
		if err != nil {
			sublogger.Error().Err(err).Send()
			return nil
		}
		defer resp.Body.Close()

		var r dohResponse
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			sublogger.Error().Err(err).Send()
			return nil
		}

		if r.Status != 0 {
			sublogger.Error().Int("status", r.Status).Send()
			return nil
		}

		answers := []dns.RR{}
		for _, ans := range r.Answer {
			record := fmt.Sprintf("%s %d %s %s", ans.Name, ans.TTL, dns.Type(ans.Type).String(), ans.Data)
			rr, err := dns.NewRR(record)
			if err != nil {
				sublogger.Error().Err(err).Send()
			}
			answers = append(answers, rr)
		}

		return answers
	}

	log.Debug().Str("module", "client.doh").Str("server", dohServer).Msg("create DOH server")
	dohClientCache.Store(serverKey, cc)
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
	Answer []struct {
		Name string `json:"name"` // The record owner.
		Type uint16 `json:"type"` // The type of DNS record.
		TTL  int    `json:"TTL"`  // The number of seconds the answer can be stored in cache before it is considered stale.
		Data string `json:"data"` // The value of the DNS record for the given name and type.
	} `json:"Answer"`
}
