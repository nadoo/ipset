# ipset

[![Go Report Card](https://goreportcard.com/badge/github.com/nadoo/ipset?style=flat-square)](https://goreportcard.com/report/github.com/nadoo/ipset)
[![GitHub tag](https://img.shields.io/github/v/tag/nadoo/ipset.svg?sort=semver&style=flat-square)](https://github.com/nadoo/ipset/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/nadoo/ipset)](https://pkg.go.dev/github.com/nadoo/ipset)

ipset package for go via netlink socket.

## Usage

### Your code:
```Go
package main

import (
	"log"

	"github.com/nadoo/ipset"
)

func main() {
	err := ipset.Init()
	if err != nil {
		log.Printf("error in create netlink: %s", err)
		return
	}

	if err = ipset.Create("myset"); err != nil {
		log.Printf("error in create set: %s", err)
		return
	}

	ipset.Add("myset", "1.1.1.1")
	ipset.Add("myset", "192.168.1.0/24")

	// ipset.Flush("myset")
}
```

### Result:
`ipset list myset`

```
Name: myset
Type: hash:net
Revision: 1
Header: family inet hashsize 1024 maxelem 65536
Size in memory: 472
References: 0
Number of entries: 2
Members:
192.168.1.0/24
1.1.1.1
```

## Links

- [glider](https://github.com/nadoo/glider): a forward proxy with ipset management features.