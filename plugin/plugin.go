package plugin

import (
	"net"
	"strings"

	"github.com/coredns/coredns/request"
	"github.com/gophercloud/gophercloud"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// OpenStack is a plugin to talk to an OpenStack API.
type OpenStack struct {
	Zone           string
	Entries        *DNSEntries
	AuthOptions    *gophercloud.AuthOptions
	Region         string
	EnableWildcard bool
}

// Name implements the plugin.Handler interface.
func (os OpenStack) Name() string {
	return "openstack"
}

// ServeDNS implements the plugin.Handler interface.
func (os OpenStack) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Compress = true
	a.Authoritative = true

	// Only answer type A queries
	if state.QType() != dns.TypeA {
		a.SetRcode(r, dns.RcodeNameError)
		state.SizeAndDo(a)
		w.WriteMsg(a)
		return 0, nil
	}

	qn := strings.TrimSuffix(state.QName(), "."+os.Zone)
	var entry []string

	if os.EnableWildcard {
		parts := strings.Split(qn, ".")
		for i := len(parts) - 2; i >= 0; i-- {
			name := strings.Join(parts[i:len(parts)], ".")
			if e := (*os.Entries)[name]; e != nil {
				entry = e
				continue
			}
		}
	} else {
		entry = (*os.Entries)[qn]
	}

	if entry != nil {
		rr := new(dns.A)
		rr.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
		rr.A = net.ParseIP(entry[0]).To4()
		a.Answer = []dns.RR{rr}
	} else {
		a.SetRcode(r, dns.RcodeNameError)
	}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return 0, nil
}
