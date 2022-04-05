package server

import (
	"strings"

	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type router struct {
	defaultRouter *routerNode
	recordRouter  map[uint16]*routerNode
}
type routerNode struct {
	next   map[string]*routerNode
	domain *routerMatched
	suffix *routerMatched
}
type routerMatched struct {
	upstream config.Upstream
	isSuffix bool
	index    int
}

///

func (r *router) search(domain string, record uint16) *config.Upstream {
	logger := log.With().
		Str("module", "server.router").
		Str("domain", domain).
		Str("record", dns.TypeToString[record]).
		Logger()

	logger.Trace().Msg("searching domain")

	segments := domainToSegments(domain)

	m1 := r.recordRouter[record].searchSegments(segments)
	m2 := r.defaultRouter.searchSegments(segments)
	if m1 != nil && m2 != nil {
		if m1.isSuffix == false {
			logger.Trace().Msg("recordRouter, domain")
			return &m1.upstream
		} else if m2.isSuffix == false {
			logger.Trace().Msg("defaultRouter, domain")
			return &m2.upstream
		}
		if m1.index < m2.index {
			logger.Trace().Msg("recordRouter, higher index")
			return &m1.upstream
		} else if m2.index < m1.index {
			logger.Trace().Msg("defaultRouter, higher index")
			return &m2.upstream
		} else {
			logger.Trace().Msg("recordRouter, record")
			return &m1.upstream
		}
	}
	if m1 != nil {
		logger.Trace().Msg("recordRouter")
		return &m1.upstream
	}
	if m2 != nil {
		logger.Trace().Msg("defaultRouter")
		return &m2.upstream
	}
	logger.Trace().Msg("not found")
	return nil
}

func (r *router) addRules(rules []*config.Rule) {
	for idx, rule := range rules {
		for _, domain := range rule.Pattern.Domain {
			r.addDomain(domain, false, rule.Pattern.Record, rule.Upstream, idx)
		}
		for _, domain := range rule.Pattern.Suffix {
			r.addDomain(domain, true, rule.Pattern.Record, rule.Upstream, idx)
		}
	}
}

func (r *router) addDomain(domain string, isSuffix bool, record string, upstream config.Upstream, idx int) {
	logger := log.With().
		Str("module", "server.router").
		Str("domain", domain).
		Bool("isSuffix", isSuffix).
		Logger()

	if len(record) > 0 {
		recordType := dns.StringToType[record]
		node, found := r.recordRouter[recordType]
		if !found {
			node = new(routerNode)
			r.recordRouter[recordType] = node
		}
		logger.Trace().
			Str("record", record).
			Msg("add domain to recordRouter")
		node.addDomain(domain, isSuffix, upstream, idx)
	} else {
		logger.Trace().Msg("add domain to defaultRouter")
		r.defaultRouter.addDomain(domain, isSuffix, upstream, idx)
	}
}

///

func (node *routerNode) searchSegments(segments []string) *routerMatched {
	if node == nil {
		return nil
	}

	curr := node
	var matched *routerMatched = curr.suffix
	for _, segment := range segments {
		if curr.next == nil {
			return matched
		}
		next, found := curr.next[segment]
		if !found {
			return matched
		}
		curr = next
		if curr.suffix != nil {
			matched = curr.suffix
		}
	}

	if curr.domain != nil {
		return curr.domain
	} else {
		return matched
	}
}

func (node *routerNode) addDomain(domain string, isSuffix bool, upstream config.Upstream, idx int) {
	segments := domainToSegments(domain)
	curr := node
	for _, segment := range segments {
		if curr.next == nil {
			curr.next = make(map[string]*routerNode)
		}
		next, found := curr.next[segment]
		if !found {
			next = new(routerNode)
			curr.next[segment] = next
		}
		curr = next
	}
	if isSuffix {
		if curr.suffix == nil || curr.suffix.index > idx {
			curr.suffix = &routerMatched{upstream, true, idx}
		}
	} else {
		if curr.domain == nil || curr.domain.index > idx {
			curr.domain = &routerMatched{upstream, false, idx}
		}
	}
}

///

func domainToSegments(domain string) []string {
	rev := []string{}
	fullDomain := dns.CanonicalName(domain)
	splited := strings.Split(fullDomain, ".")
	for i := len(splited) - 2; i >= 0; i-- {
		if len(splited[i]) > 0 {
			rev = append(rev, splited[i])
		}
	}
	return rev
}
