package ipset

import (
	"github.com/nadoo/ipset/internal/netlink"
)

var nl *netlink.NetLink

// Create creates a new set.
func Create(setName string) (err error) {
	if nl == nil {
		if nl, err = netlink.New(); err != nil {
			return
		}
	}
	return nl.CreateSet(setName)
}

// Flush flushes a named set.
func Flush(setName string) (err error) {
	if nl == nil {
		if nl, err = netlink.New(); err != nil {
			return
		}
	}
	return nl.FlushSet(setName)
}

// Add adds entry to the named set.
func Add(setName, entry string) (err error) {
	if nl == nil {
		if nl, err = netlink.New(); err != nil {
			return
		}
	}
	return nl.AddToSet(setName, entry)
}
