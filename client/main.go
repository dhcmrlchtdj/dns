package client

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

type Answer struct {
	// The record owner.
	Name string `json:"name"`
	// The type of DNS record.
	Type uint16 `json:"type"`
	// The number of seconds the answer can be stored in cache before it is considered stale.
	TTL int `json:"TTL"`
	// The value of the DNS record for the given name and type.
	Data string `json:"data"`
}

type DNSClient func(string, uint16) []Answer

func Query(name string, qtype uint16) []Answer {
	client := dohClientPool.Get().(*http.Client)
	defer func() {
		dohClientPool.Put(client)
	}()

	req, err := http.NewRequest("GET", "https://cloudflare-dns.com/dns-query", nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	req.Header.Set("accept", "application/dns-json")
	q := req.URL.Query()
	q.Set("name", name)
	q.Set("type", dns.Type(qtype).String())
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	var r DNSResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Println(err)
		return nil
	}
	return r.Answer
}
