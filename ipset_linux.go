package ipset

import (
	"net/netip"

	"github.com/nadoo/ipset/internal/netlink"
)

var nl *netlink.NetLink

// Option is used to set parameters of ipset operations.
type Option = netlink.Option

// OptIPv6 sets `family inet6` parameter to operations.
func OptIPv6() Option { return func(opts *netlink.Options) { opts.IPv6 = true } }

// OptTimeout sets `timeout xx` parameter to operations.
func OptTimeout(timeout uint32) Option { return func(opts *netlink.Options) { opts.Timeout = timeout } }

// Init prepares a netlink socket of ipset.
func Init() (err error) {
	nl, err = netlink.New()
	return err
}

// Create creates a new set.
func Create(setName string, opts ...Option) (err error) {
	return nl.CreateSet(setName, opts...)
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
// entry could be: "1.1.1.1" or "192.168.1.0/24" or "2022::1" or "2022::1/32".
func Add(setName, entry string, opts ...Option) (err error) {
	return handleEntry(netlink.IPSET_CMD_ADD, setName, entry, opts...)
}

// Del deletes an entry from the named set.
// entry could be: "1.1.1.1" or "192.168.1.0/24" or "2022::1" or "2022::1/32".
func Del(setName, entry string) (err error) {
	return handleEntry(netlink.IPSET_CMD_DEL, setName, entry)
}

func handleEntry(cmd int, setName, entry string, opts ...Option) error {
	ip, err := netip.ParseAddr(entry)
	if err == nil {
		return nl.HandleAddr(cmd, setName, ip, netip.Prefix{}, opts...)
	}
	cidr, err := netip.ParsePrefix(entry)
	if err == nil {
		return nl.HandleAddr(cmd, setName, cidr.Addr(), cidr, opts...)
	}
	return err
}

// AddAddr adds an addr to the named set.
func AddAddr(setName string, ip netip.Addr, opts ...Option) (err error) {
	return nl.HandleAddr(netlink.IPSET_CMD_ADD, setName, ip, netip.Prefix{}, opts...)
}

// DelAddr deletes an addr from the named set.
func DelAddr(setName string, ip netip.Addr) (err error) {
	return nl.HandleAddr(netlink.IPSET_CMD_DEL, setName, ip, netip.Prefix{})
}

// AddPrefix adds a cidr to the named set.
func AddPrefix(setName string, cidr netip.Prefix, opts ...Option) (err error) {
	return nl.HandleAddr(netlink.IPSET_CMD_ADD, setName, cidr.Addr(), cidr, opts...)
}

// DelPrefix deletes a cidr from the named set.
func DelPrefix(setName string, cidr netip.Prefix) (err error) {
	return nl.HandleAddr(netlink.IPSET_CMD_DEL, setName, cidr.Addr(), cidr)
}
