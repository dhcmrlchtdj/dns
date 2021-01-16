package client

import (
	"strings"
)

type dnsRouter struct {
	matched dnsClient
	router  map[string]*dnsRouter
}

func (c *dnsRouter) add(domain string, cli dnsClient) {
	if domain == "." {
		if c.matched == nil {
			c.matched = cli
		}
	} else {
		r := c
		for _, part := range revDomain(domain) {
			if r.router == nil {
				r.router = make(map[string]*dnsRouter)
			}
			next, found := r.router[part]
			if !found {
				next = new(dnsRouter)
				r.router[part] = next
			}
			r = next
		}
		if r.matched == nil {
			r.matched = cli
		}
	}
}

func (c *dnsRouter) route(domain string) dnsClient {
	if domain == "." {
		return c.matched
	} else {
		matched := c.matched

		r := c
		for _, part := range revDomain(domain) {
			if r.router == nil {
				break
			} else {
				next, found := r.router[part]
				if found {
					if next.matched != nil {
						matched = next.matched
					}
					r = next
				} else {
					break
				}
			}
		}

		return matched
	}
}

func revDomain(domain string) []string {
	rev := []string{}
	splited := strings.Split(domain, ".")
	for i := len(splited) - 2; i >= 0; i-- {
		rev = append(rev, splited[i])
	}
	return rev
}
