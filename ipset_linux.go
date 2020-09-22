package ipset

import (
	"github.com/nadoo/ipset/internal/netlink"
)

var nl *netlink.NetLink

// Init initializes a netlink socket.
func Init() (err error) {
	nl, err = netlink.New()
	return err
}

// Create creates a new set.
func Create(setName string) error {
	return nl.CreateSet(setName)
}

// Flush flushes a named set.
func Flush(setName string) error {
	return nl.FlushSet(setName)
}

// Add adds entry to the named set.
func Add(setName, entry string) error {
	return nl.AddToSet(setName, entry)
}
