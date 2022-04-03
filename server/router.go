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
		logger.Trace().Msg("found on recordRouter and defaultRouter")
		if m1.isSuffix == false {
			return &m1.upstream
		}
		if m2.isSuffix == false {
			return &m2.upstream
		}
		return &m1.upstream
	}
	if m1 != nil {
		logger.Trace().Msg("found on recordRouter")
		return &m1.upstream
	}
	if m2 != nil {
		logger.Trace().Msg("found on defaultRouter")
		return &m2.upstream
	}
	logger.Trace().Msg("not found")
	return nil
}

func (r *router) addRules(rules []config.Rule) {
	for _, rule := range rules {
		for _, domain := range rule.Pattern.Domain {
			r.addDomain(domain, false, rule.Pattern.Record, rule.Upstream)
		}
		for _, domain := range rule.Pattern.Suffix {
			r.addDomain(domain, true, rule.Pattern.Record, rule.Upstream)
		}
	}
}

func (r *router) addDomain(domain string, isSuffix bool, record string, upstream config.Upstream) {
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
		node.addDomain(domain, isSuffix, upstream)
	} else {
		logger.Trace().Msg("add domain to defaultRouter")
		r.defaultRouter.addDomain(domain, isSuffix, upstream)
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

func (node *routerNode) addDomain(domain string, isSuffix bool, upstream config.Upstream) {
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
		if curr.suffix == nil {
			curr.suffix = &routerMatched{upstream, true}
		}
	} else {
		if curr.domain == nil {
			curr.domain = &routerMatched{upstream, false}
		}
	}
}

///

func domainToSegments(domain string) []string {
	rev := []string{}
	fullDomain := dns.Fqdn(domain)
	splited := strings.Split(fullDomain, ".")
	for i := len(splited) - 2; i >= 0; i-- {
		if len(splited[i]) > 0 {
			rev = append(rev, splited[i])
		}
	}
	return rev
}
