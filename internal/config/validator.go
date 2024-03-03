package config

import (
	"net"
	"net/url"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
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
	ErrPatternInvalid      = errors.New("invalid pattern")
	ErrPatternDomain       = errors.New("invalid domain pattern")
	ErrPatternRecord       = errors.New("invalid record type")
	ErrPatternBuiltin      = errors.New("invalid builtin rule")
	ErrPatternBuiltinProxy = errors.New("invalid builtin proxy")
)

func (pat *Pattern) IsValid() error {
	if pat == nil {
		return ErrPatternInvalid
	}
	if pat.Builtin != "" {
		switch pat.Builtin {
		case "china-list": // do nothing
		default:
			return errors.Wrap(ErrPatternBuiltin, pat.Builtin)
		}
		if pat.BuiltinProxy != "" {
			if _, err := url.Parse(pat.BuiltinProxy); err != nil {
				return errors.Wrap(ErrPatternBuiltinProxy, pat.BuiltinProxy)
			}
		}
	} else if len(pat.Domain) == 0 && len(pat.Suffix) == 0 {
		return ErrPatternDomain
	}
	if pat.Record != "" {
		if _, found := dns.StringToType[pat.Record]; !found {
			return errors.Wrap(ErrPatternRecord, pat.Record)
		}
	}
	return nil
}

var (
	ErrUpstreamInvalid     = errors.New("invalid upstream")
	ErrUpstreamBlockAction = errors.New("unsupported block action")
	ErrUpstreamIpv4        = errors.New("invalid IPv4")
	ErrUpstreamIpv6        = errors.New("invalid IPv6")
	ErrUpstreamUdp         = errors.New("invalid UDP")
	ErrUpstreamDoh         = errors.New("invalid DOH")
	ErrUpstreamDohProxy    = errors.New("invalid DOH proxy")
)

func (up *Upstream) IsValid() error {
	if up == nil {
		return ErrUpstreamInvalid
	}
	if up.Block != "" {
		if up.Block != "nodata" && up.Block != "nxdomain" {
			return errors.Wrap(ErrUpstreamBlockAction, up.Block)
		}
		if up.Ipv4 != "" || up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Ipv4 != "" {
		if net.ParseIP(up.Ipv4) == nil || strings.Contains(up.Ipv4, ":") {
			return errors.Wrap(ErrUpstreamIpv4, up.Ipv4)
		}
		if up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Ipv6 != "" {
		if net.ParseIP(up.Ipv6) == nil || strings.Count(up.Ipv6, ":") < 2 {
			return errors.Wrap(ErrUpstreamIpv6, up.Ipv6)
		}
		if up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Udp != "" {
		if _, _, err := net.SplitHostPort(up.Udp); err != nil {
			return errors.Wrap(ErrUpstreamUdp, up.Udp)
		}
		if up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Doh != "" {
		if _, err := url.Parse(up.Doh); err != nil {
			return errors.Wrap(ErrUpstreamDoh, up.Doh)
		}
	}
	if up.DohProxy != "" {
		if _, err := url.Parse(up.DohProxy); err != nil {
			return errors.Wrap(ErrUpstreamDohProxy, up.DohProxy)
		}
		if up.Doh == "" {
			return ErrUpstreamInvalid
		}
	}
	return nil
}
