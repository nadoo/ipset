# ipset

[![Go Report Card](https://goreportcard.com/badge/github.com/nadoo/ipset?style=flat-square)](https://goreportcard.com/report/github.com/nadoo/ipset)
[![GitHub tag](https://img.shields.io/github/v/tag/nadoo/ipset.svg?sort=semver&style=flat-square)](https://github.com/nadoo/ipset/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/nadoo/ipset)](https://pkg.go.dev/github.com/nadoo/ipset)

netlink ipset package for Go.

## Usage

### Your code:
```Go
package main

import (
	"log"

	"github.com/nadoo/ipset"
)

func main() {
	// must call Init first
	if err := ipset.Init(); err != nil {
		log.Printf("error in ipset Init: %s", err)
		return
	}

	// default is ipv4
	ipset.Destroy("myset")
	ipset.Create("myset")

	ipset.Add("myset", "1.1.1.1")
	ipset.Flush("myset")

	ipset.Add("myset", "192.168.1.0/24")
	ipset.Del("myset", "192.168.1.0/24")

	ipset.Add("myset", "2.2.2.2")

	// ipset family inet6
	ipset.Destroy("myset6")
	ipset.Create("myset6", ipset.OptIPv6(), ipset.OptTimeout(60))
	ipset.Flush("myset6")

	ipset.Add("myset6", "2404:6800:4005:812::200e", ipset.OptTimeout(10))
	ipset.Add("myset6", "2404:6800:4005:812::/64")
}
```

### Result:
`ipset list myset`

```
Name: myset
Type: hash:net
Revision: 1
Header: family inet hashsize 1024 maxelem 65536
Size in memory: 408
References: 0
Number of entries: 1
Members:
2.2.2.2
```

`ipset list myset6`

```
Name: myset6
Type: hash:net
Revision: 1
Header: family inet6 hashsize 1024 maxelem 65536 timeout 60
Size in memory: 1432
References: 0
Number of entries: 2
Members:
2404:6800:4005:812::/64 timeout 55
2404:6800:4005:812::200e timeout 5
```

## Links

- [glider](https://github.com/nadoo/glider): a forward proxy with ipset management features powered by this package.