package server

import (
	"context"
	"strings"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/config"
)

type router struct {
	domain                 *routerNode
	domainSuffix           *routerNode
	domainWithRecord       map[uint16]*routerNode
	domainSuffixWithRecord map[uint16]*routerNode
}
type routerNode struct {
	next    map[string]*routerNode
	matched *routerMatched
}
type routerMatched struct {
	priority int // smaller means higher priority
	upstream *config.Upstream
}

///

func (r *router) setup() {
	r.domain = new(routerNode)
	r.domainSuffix = new(routerNode)
	r.domainWithRecord = make(map[uint16]*routerNode)
	r.domainSuffixWithRecord = make(map[uint16]*routerNode)
}

func (r *router) addRules(ctx context.Context, rules []*config.Rule) {
	for priority, rule := range rules {
		for _, domain := range rule.Pattern.Domain {
			r.addDomain(ctx, priority, domain, false, rule.Pattern.Record, &rule.Upstream)
		}
		for _, domain := range rule.Pattern.Suffix {
			r.addDomain(ctx, priority, domain, true, rule.Pattern.Record, &rule.Upstream)
		}
	}
}

func (r *router) addDomain(
	ctx context.Context,
	priority int,
	domain string,
	isSuffix bool,
	record string,
	upstream *config.Upstream,
) {
	zerolog.Ctx(ctx).
		Trace().
		Str("module", "server.router").
		Int("priority", priority).
		Str("domain", domain).
		Bool("isSuffix", isSuffix).
		Str("record", record).
		Msg("added")

	if len(record) > 0 {
		recordRouter := r.domainWithRecord
		if isSuffix {
			recordRouter = r.domainSuffixWithRecord
		}

		recordType := dns.StringToType[record]
		node, found := recordRouter[recordType]
		if !found {
			node = new(routerNode)
			recordRouter[recordType] = node
		}

		node.addDomain(priority, domain, upstream)
	} else {
		if isSuffix {
			r.domainSuffix.addDomain(priority, domain, upstream)
		} else {
			r.domain.addDomain(priority, domain, upstream)
		}
	}
}

func (r *router) search(ctx context.Context, domain string, record uint16) *config.Upstream {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.router").
		Str("domain", domain).
		Str("record", dns.TypeToString[record]).
		Logger()

	segments := domainToSegments(domain)

	c1, m1 := r.domainWithRecord[record].searchSegments(segments)
	if m1 != nil && c1 == len(segments) {
		logger.Trace().Dict("match", zerolog.Dict().Bool("record", true).Bool("suffix", false).Int("priority", m1.priority)).Bool("found", true).Send()
		return m1.upstream
	}

	c2, m2 := r.domain.searchSegments(segments)
	if m2 != nil && c2 == len(segments) {
		logger.Trace().Dict("match", zerolog.Dict().Bool("record", false).Bool("suffix", false).Int("priority", m2.priority)).Bool("found", true).Send()
		return m2.upstream
	}

	c3, m3 := r.domainSuffixWithRecord[record].searchSegments(segments)
	c4, m4 := r.domainSuffix.searchSegments(segments)
	if m3 != nil && m4 != nil {
		if c3 > c4 {
			logger.Trace().Dict("match", zerolog.Dict().Bool("record", true).Bool("suffix", true).Int("priority", m3.priority)).Bool("found", true).Send()
			return m3.upstream
		} else if c3 < c4 {
			logger.Trace().Dict("match", zerolog.Dict().Bool("record", false).Bool("suffix", true).Int("priority", m4.priority)).Bool("found", true).Send()
			return m4.upstream
		} else if m3.priority <= m4.priority {
			logger.Trace().Dict("match", zerolog.Dict().Bool("record", true).Bool("suffix", true).Int("priority", m3.priority)).Bool("found", true).Send()
			return m3.upstream
		} else {
			logger.Trace().Dict("match", zerolog.Dict().Bool("record", false).Bool("suffix", true).Int("priority", m4.priority)).Bool("found", true).Send()
			return m4.upstream
		}
	} else if m3 != nil {
		logger.Trace().Dict("match", zerolog.Dict().Bool("record", true).Bool("suffix", true).Int("priority", m3.priority)).Bool("found", true).Send()
		return m3.upstream
	} else if m4 != nil {
		logger.Trace().Dict("match", zerolog.Dict().Bool("record", false).Bool("suffix", true).Int("priority", m4.priority)).Bool("found", true).Send()
		return m4.upstream
	}

	logger.Trace().Bool("found", false).Send()
	return nil
}

///

func (node *routerNode) searchSegments(segments []string) (int, *routerMatched) {
	if node == nil {
		return 0, nil
	}

	curr := node
	longestMatch := 0
	segmentCount := 0
	var matched *routerMatched = curr.matched
	for _, segment := range segments {
		if curr.next == nil {
			break
		}
		next, found := curr.next[segment]
		if !found {
			break
		}
		curr = next
		segmentCount++
		if curr.matched != nil {
			longestMatch = segmentCount
			matched = curr.matched
		}
	}
	return longestMatch, matched
}

func (node *routerNode) addDomain(priority int, domain string, upstream *config.Upstream) {
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
	if curr.matched == nil || curr.matched.priority > priority {
		curr.matched = &routerMatched{priority, upstream}
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
