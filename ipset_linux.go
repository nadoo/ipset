package ipset

import (
	"github.com/nadoo/ipset/internal/netlink"
)

var nl *netlink.NetLink

// Init prepares a netlink socket of ipset.
func Init() (err error) {
	nl, err = netlink.New()
	return err
}

// Create creates a new set.
func Create(setName string) (err error) {
	return nl.CreateSet(setName)
}

// Destroy destroys a named set.
func Destroy(setName string) (err error) {
	return nl.DestroySet(setName)
}

// Flush flushes a named set.
func Flush(setName string) (err error) {
	return nl.FlushSet(setName)
}

// Add adds an entry to the named set.
// entry could be: "1.1.1.1" or "192.168.1.0/24".
func Add(setName, entry string) (err error) {
	return nl.AddToSet(setName, entry)
}

// Del deletes an entry from the named set.
func Del(setName, entry string) (err error) {
	return nl.DelFromSet(setName, entry)
}
