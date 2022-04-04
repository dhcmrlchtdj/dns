package config

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/miekg/dns"
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
	ErrPatternInvalid = errors.New("invalid pattern")
	ErrPatternDomain  = errors.New("both domain/suffix are empty")
	ErrPatternRecord  = errors.New("invalid record")
)

func (pat *Pattern) IsValid() error {
	if pat == nil {
		return ErrPatternInvalid
	}
	if len(pat.Domain) == 0 && len(pat.Suffix) == 0 {
		return ErrPatternDomain
	}
	if len(pat.Record) > 0 {
		if _, found := dns.StringToType[pat.Record]; !found {
			return ErrPatternRecord
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
			return ErrUpstreamBlockAction
		}
		if up.Ipv4 != "" || up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Ipv4 != "" {
		if net.ParseIP(up.Ipv4) == nil || strings.Contains(up.Ipv4, ":") {
			return ErrUpstreamIpv4
		}
		if up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Ipv6 != "" {
		if net.ParseIP(up.Ipv6) == nil || strings.Count(up.Ipv6, ":") < 2 {
			return ErrUpstreamIpv6
		}
		if up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Udp != "" {
		if _, _, err := net.SplitHostPort(up.Udp); err != nil {
			return ErrUpstreamUdp
		}
		if up.Doh != "" || up.DohProxy != "" {
			return ErrUpstreamInvalid
		}
	}
	if up.Doh != "" {
		if _, err := url.Parse(up.Doh); err != nil {
			return ErrUpstreamDoh
		}
	}
	if up.DohProxy != "" {
		if _, err := url.Parse(up.DohProxy); err != nil {
			return ErrUpstreamDohProxy
		}
		if up.Doh == "" {
			return ErrUpstreamInvalid
		}
	}
	return nil
}
