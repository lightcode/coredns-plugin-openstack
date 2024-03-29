:warning: **This project is not maintained anymore.**

---

# openstack

## Name

*openstack* — adds the ability to convert server names into their floating IPs.

## Description

This plugin allows resolving names build like `<server_name>.<project_name>` into the
corresponding floating IP.

## Syntax

```
openstack {
    auth_url AUTHENTICATION_URL
    username USERNAME
    passwork PASSWORD
    domain_name DOMAIN_NAME
    region REGION_NAME
    wildcard
}
```

* `auth_url` specifies the Keystone authentication URL. *Required*.
* `username` specifies the name of a user who can list tenants and list all servers. Defaults to `coredns`.
* `password` specifies the password of the user. *Required*.
* `domain_name` specifies the Keystone domain which the user belongs to. Defaults to `default`.
* `region` specifies the OpenStack region for your servers. Defaults to `RegionOne`.
* `wildcard` set this option to resolve every names that end by `<server_name>.<project_name>`.

## Examples

```
openstack {
    openstack {
        auth_url "http://your.keystone.endoint/v3"
        username "coredns"
        passwork "SET HERE YOUR PASSWORD"
        domain_name "default"
        region "RegionOne"
        wildcard
    }
    errors
    log
}
```
