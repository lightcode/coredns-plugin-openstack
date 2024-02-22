package plugin

import (
	"net"
	"strings"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/rackspace/gophercloud"
	"golang.org/x/net/context"
)

// OpenStack is a plugin to talk to an OpenStack API.
type OpenStack struct {
	Zone        string
	Entries     *DNSEntries
	AuthOptions *gophercloud.AuthOptions
	Region      string
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

	qn := strings.TrimSuffix(state.QName(), os.Zone)
	qn = strings.TrimSuffix(qn, ".")

	if entry := (*os.Entries)[qn]; entry != nil {
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
