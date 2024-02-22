package plugin

import (
	"fmt"
	"log"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/pagination"
)

// DNSEntries represents a list of hostname assiociated with a list of IPs
type DNSEntries map[string][]string

type serverEntry struct {
	Name       string
	TenantName string
	Addresses  []string
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

func (os OpenStack) listTenants(provider *gophercloud.ProviderClient) (map[string]string, error) {
	r := make(map[string]string)

	client, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{Region: os.Region})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch tenants: %s", err)
	}

	pager := projects.List(client, nil)
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		projectsList, _ := projects.ExtractProjects(page)

		for _, t := range projectsList {
			r[t.ID] = t.Name
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to fetch tenants: %s", err)
	}

	return r, nil
}

func (os OpenStack) listServers(provider *gophercloud.ProviderClient, tenantMapping map[string]string) ([]serverEntry, error) {
	serverEntries := make([]serverEntry, 0)

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: os.Region,
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

func (os OpenStack) fetchEntries() (DNSEntries, error) {
	provider, err := openstack.AuthenticatedClient(*os.AuthOptions)
	if err != nil {
		return nil, fmt.Errorf("Unable to authenticate: %s", err)
	}

	entries := make(DNSEntries)
	mapping, err := os.listTenants(provider)
	if err != nil {
		return nil, err
	}

	servers, err := os.listServers(provider, mapping)
	if err != nil {
		return nil, err
	}

	for _, srv := range servers {
		name := srv.Name + "." + srv.TenantName
		entries[name] = srv.Addresses
	}

	return entries, nil
}

func (os OpenStack) runFetchEntries(globalEntries *DNSEntries) {
	var err error
	var entriesTemp DNSEntries

	for {
		entriesTemp, err = os.fetchEntries()
		if err == nil {
			*globalEntries = entriesTemp
		} else {
			log.Println(plugin.Error("openstack", err))
		}
		time.Sleep(time.Second)
	}
}
