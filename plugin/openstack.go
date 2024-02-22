package plugin

import (
	"fmt"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/identity/v2/tenants"
	"github.com/rackspace/gophercloud/pagination"
)

// DNSEntries represents a list of hostname assiociated with a list of IPs
type DNSEntries map[string][]string

type serverEntry struct {
	Name       string
	TenantName string
	Addresses  []string
}

func listTenants(provider *gophercloud.ProviderClient) (map[string]string, error) {
	r := make(map[string]string)

	client := openstack.NewIdentityV2(provider)
	pager := tenants.List(client, nil)
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		tenantList, _ := tenants.ExtractTenants(page)

		for _, t := range tenantList {
			r[t.ID] = t.Name
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to fetch tenants: %s", err)
	}

	return r, nil
}

func listServers(provider *gophercloud.ProviderClient, tenantMapping map[string]string, region string) ([]serverEntry, error) {
	serverEntries := make([]serverEntry, 0)

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch servers: %s", err)
	}

	opts := servers.ListOpts{AllTenants: true}
	pager := servers.List(client, opts)
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, _ := servers.ExtractServers(page)

		for _, s := range serverList {
			se := serverEntry{
				Name:       s.Name,
				TenantName: tenantMapping[s.TenantID],
				Addresses:  extractFloating(s.Addresses),
			}
			serverEntries = append(serverEntries, se)
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to list servers: %s", err)
	}

	return serverEntries, nil
}

func extractFloating(addresses map[string]interface{}) []string {
	floatings := make([]string, 0)
	for _, addrList := range addresses {
		t := addrList.([]interface{})
		for _, j := range t {
			k := j.(map[string]interface{})
			if k["OS-EXT-IPS:type"] == "floating" && k["version"].(float64) == 4 {
				floatings = append(floatings, k["addr"].(string))
			}
		}
	}
	return floatings
}

func fetchEntries(authOpts *gophercloud.AuthOptions, region string) (DNSEntries, error) {
	provider, err := openstack.AuthenticatedClient(*authOpts)
	if err != nil {
		return nil, fmt.Errorf("Unable to authenticate: %s", err)
	}

	entries := make(DNSEntries)
	mapping, err := listTenants(provider)
	if err != nil {
		return nil, err
	}

	servers, err := listServers(provider, mapping, region)
	if err != nil {
		return nil, err
	}

	for _, srv := range servers {
		name := srv.Name + "." + srv.TenantName
		entries[name] = srv.Addresses
	}

	return entries, nil
}
