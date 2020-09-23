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

// Add adds an entry to the named set.
// entry could be: "1.1.1.1" or "192.168.1.0/24".
func Add(setName, entry string) (err error) {
	if nl == nil {
		if nl, err = netlink.New(); err != nil {
			return
		}
	}
	return nl.AddToSet(setName, entry)
}

// Del deletes an entry from the named set.
func Del(setName, entry string) (err error) {
	if nl == nil {
		if nl, err = netlink.New(); err != nil {
			return
		}
	}
	return nl.DelFromSet(setName, entry)
}
