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

func (pat *Pattern) IsValid() error {
	if pat == nil {
		return errors.New("nil pattern")
	}
	if len(pat.Domain) == 0 && len(pat.Suffix) == 0 {
		return errors.New("both domain/suffix are empty in the pattern")
	}
	if len(pat.Record) > 0 {
		if _, found := dns.StringToType[pat.Record]; !found {
			return errors.New("invalid record type")
		}
	}
	return nil
}

func (up *Upstream) IsValid() error {
	if up == nil {
		return errors.New("nil upstream")
	}
	if up.Block != "" {
		if up.Block != "nodata" && up.Block != "nxdomain" {
			return errors.New("unsupported block action")
		}
		if up.Ipv4 != "" || up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return errors.New("invalid upstream")
		}
	}
	if up.Ipv4 != "" {
		if net.ParseIP(up.Ipv4) == nil || strings.Contains(up.Ipv4, ":") {
			return errors.New("invalid IPv4 address")
		}
		if up.Ipv6 != "" || up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return errors.New("invalid upstream")
		}
	}
	if up.Ipv6 != "" {
		if net.ParseIP(up.Ipv6) == nil || strings.Count(up.Ipv6, ":") < 2 {
			return errors.New("invalid IPv6 address")
		}
		if up.Udp != "" || up.Doh != "" || up.DohProxy != "" {
			return errors.New("invalid upstream")
		}
	}
	if up.Udp != "" {
		if _, _, err := net.SplitHostPort(up.Udp); err != nil {
			return errors.New("invalid UDP address")
		}
		if up.Doh != "" || up.DohProxy != "" {
			return errors.New("invalid upstream")
		}
	}
	if up.Doh != "" {
		if _, err := url.Parse(up.Doh); err != nil {
			return errors.New("invalid DoH address")
		}
	}
	if up.DohProxy != "" {
		if _, err := url.Parse(up.DohProxy); err != nil {
			return errors.New("invalid proxy")
		}
		if up.Doh == "" {
			return errors.New("invalid upstream")
		}
	}
	return nil
}
