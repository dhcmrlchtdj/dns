package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

var dohClientCache = new(sync.Map)

type Doh struct {
	server     string
	httpClient *http.Client
}

func createDohResolver(upstream *config.Upstream) *Doh {
	logger := log.With().
		Str("module", "client.doh").
		Logger()

	cacheKey := upstream.Doh + "|" + upstream.DohProxy
	if client, found := dohClientCache.Load(cacheKey); found {
		logger.Trace().Msg("get resolver from cache")
		return client.(*Doh)
	} else {
		httpClient := new(http.Client)
		if len(upstream.DohProxy) > 0 {
			proxyUrl, err := url.Parse(upstream.DohProxy)
			if err != nil {
				panic(err)
			}
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
		}
		client := &Doh{server: upstream.Doh, httpClient: httpClient}
		dohClientCache.Store(cacheKey, client)
		logger.Trace().Msg("new resolver created")
		return client
	}
}

func (s *Doh) Resolve(question dns.Question) ([]dns.RR, error) {
	logger := log.With().
		Str("module", "client.doh").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	req, err := http.NewRequest("GET", s.server, http.NoBody)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create request")
		return nil, err
	}

	req.Header.Set("accept", "application/dns-json")
	q := req.URL.Query()
	q.Set("name", question.Name)                    // Query Name
	q.Set("type", dns.TypeToString[question.Qtype]) // Query Type
	// q.Set("do", "true")                             // DO bit - set if client wants DNSSEC data
	// q.Set("cd", "false")                            // CD bit - set to disable validation
	req.URL.RawQuery = q.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to send request")
		return nil, err
	}
	defer resp.Body.Close()

	var r dohResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		logger.Error().Err(err).Msg("failed to parse response")
		return nil, err
	}

	if r.Status != 0 {
		logger.Trace().
			Str("rcode", dns.RcodeToString[r.Status]).
			Msg("failed to resolved")
		return nil, &ErrDnsResponse{Rcode: r.Status}
	}

	answers := []dns.RR{}
	for _, ans := range r.Answer {
		record := fmt.Sprintf(
			"%s %d %s %s",
			ans.Name,
			ans.TTL,
			dns.TypeToString[ans.Type],
			ans.Data,
		)
		rr, err := dns.NewRR(record)
		if err != nil {
			logger.Error().Err(err).Send()
			return nil, err
		}
		answers = append(answers, rr)
	}

	logger.Trace().Msg("resolved")
	return answers, nil
}

// https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-https/make-api-requests/dns-json/
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
