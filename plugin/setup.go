package plugin

import (
	"log"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mholt/caddy"
	"github.com/rackspace/gophercloud"
)

func init() {
	caddy.RegisterPlugin("openstack", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	os, err := openstackParse(c)
	if err != nil {
		return plugin.Error("openstack", err)
	}

	config := dnsserver.GetConfig(c)
	os.Zone = config.Zone

	c.OnStartup(func() error {
		go runFetchEntries(os.Entries, os.AuthOptions)
		return nil
	})

	config.AddPlugin(func(next plugin.Handler) plugin.Handler {
		return os
	})
	return nil
}

func runFetchEntries(globalEntries *DNSEntries, authOpts *gophercloud.AuthOptions) {
	var err error
	var entriesTemp DNSEntries

	for {
		entriesTemp, err = fetchEntries(authOpts)
		if err == nil {
			*globalEntries = entriesTemp
		} else {
			log.Println(plugin.Error("openstack", err))
		}
		time.Sleep(time.Second)
	}
}

func openstackParse(c *caddy.Controller) (*OpenStack, error) {
	entries := make(DNSEntries)
	authOpts := gophercloud.AuthOptions{
		Username:   "coredns",
		DomainName: "default",
	}

	os := OpenStack{}
	os.Entries = &entries
	os.AuthOptions = &authOpts

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "auth_url":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				authOpts.IdentityEndpoint = args[0]
			case "username":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				authOpts.Username = args[0]
			case "password":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				authOpts.Password = args[0]
			case "domain_name":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				authOpts.DomainName = args[0]
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}

	return &os, nil
}
