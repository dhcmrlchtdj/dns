package config

import (
	"net"
	"net/url"
	"strings"

	"github.com/miekg/dns"
	"github.com/morikuni/failure"
)

func (r *Rule) IsValid() error {
	if r == nil {
		return nil
	}
	if err := r.Pattern.IsValid(); err != nil {
		return err
	}
	if err := r.Upstream.IsValid(); err != nil {
		return err
	}
	// TODO: ipv4 can't use without record A
	return nil
}

var (
	ErrPatternInvalid      failure.StringCode = "InvalidPattern"
	ErrPatternDomain       failure.StringCode = "InvalidDomain"
	ErrPatternRecord       failure.StringCode = "InvalidRecord"
	ErrPatternBuiltin      failure.StringCode = "invalidBuiltInRule"
	ErrPatternBuiltinProxy failure.StringCode = "InvalidBuiltInProxy"
)

func (pat *Pattern) IsValid() error {
	if pat == nil {
		return failure.New(ErrPatternInvalid)
	}
	if pat.Builtin != "" {
		switch pat.Builtin {
		case "china-list": // do nothing
		default:
			return failure.New(ErrPatternBuiltin)
		}
		if pat.BuiltinProxy != "" {
			if _, err := url.Parse(pat.BuiltinProxy); err != nil {
				return failure.New(ErrPatternBuiltinProxy)
			}
		}
	} else if len(pat.Domain) == 0 && len(pat.Suffix) == 0 {
		return failure.New(ErrPatternDomain)
	}
	if pat.Record != "" {
		if _, found := dns.StringToType[pat.Record]; !found {
			return failure.New(ErrPatternRecord)
		}
	}
	return nil
}

var (
	ErrUpstreamInvalid     failure.StringCode = "InvalidUpstream"
	ErrUpstreamBlockAction failure.StringCode = "UnsupportedBlockAction"
	ErrUpstreamIpv4        failure.StringCode = "InvalidIpv4"
	ErrUpstreamIpv6        failure.StringCode = "InvalidIpv6"
	ErrUpstreamUdp         failure.StringCode = "InvalidUdp"
	ErrUpstreamDoh         failure.StringCode = "InvalidDoh"
	ErrUpstreamDohProxy    failure.StringCode = "InvalidDohProxy"
)

func (up *Upstream) IsValid() error {
	if up == nil {
		return failure.New(ErrUpstreamInvalid)
	}
	if up.Block != "" {
		if up.Block != "nodata" && up.Block != "nxdomain" {
			return failure.New(ErrUpstreamBlockAction)
		}
		if up.Ipv4 != "" || up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return failure.New(ErrUpstreamInvalid)
		}
	}
	if up.Ipv4 != "" {
		if net.ParseIP(up.Ipv4) == nil || strings.Contains(up.Ipv4, ":") {
			return failure.New(ErrUpstreamIpv4)
		}
		if up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return failure.New(ErrUpstreamInvalid)
		}
	}
	if up.Ipv6 != "" {
		if net.ParseIP(up.Ipv6) == nil || strings.Count(up.Ipv6, ":") < 2 {
			return failure.New(ErrUpstreamIpv6)
		}
		if up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return failure.New(ErrUpstreamInvalid)
		}
	}
	if up.Udp != "" {
		if _, _, err := net.SplitHostPort(up.Udp); err != nil {
			return failure.New(ErrUpstreamUdp)
		}
		if up.Doh != "" || up.DohProxy != "" {
			return failure.New(ErrUpstreamInvalid)
		}
	}
	if up.Doh != "" {
		if _, err := url.Parse(up.Doh); err != nil {
			return failure.New(ErrUpstreamDoh)
		}
	}
	if up.DohProxy != "" {
		if _, err := url.Parse(up.DohProxy); err != nil {
			return failure.New(ErrUpstreamDohProxy)
		}
		if up.Doh == "" {
			return failure.New(ErrUpstreamInvalid)
		}
	}
	return nil
}
