# ipset

[![Go Report Card](https://goreportcard.com/badge/github.com/nadoo/ipset?style=flat-square)](https://goreportcard.com/report/github.com/nadoo/ipset)
[![GitHub tag](https://img.shields.io/github/v/tag/nadoo/ipset.svg?sort=semver&style=flat-square)](https://github.com/nadoo/ipset/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/nadoo/ipset)](https://pkg.go.dev/github.com/nadoo/ipset)

ipset package for Go via netlink socket.

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

	ipset.Create("myset")

	ipset.Add("myset", "1.1.1.1")
	ipset.Flush("myset")

	ipset.Add("myset", "192.168.1.0/24")
	ipset.Del("myset", "192.168.1.0/24")

	ipset.Add("myset", "2.2.2.2")
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

## Links

- [glider](https://github.com/nadoo/glider): a forward proxy with ipset management features powered by this package.