package netlink

import (
	"encoding/binary"
	"errors"
	"net/netip"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// NFNL_SUBSYS_IPSET is netfilter netlink message type of ipset.
// https://github.com/torvalds/linux/blob/9e66317d3c92ddaab330c125dfe9d06eee268aff/include/uapi/linux/netfilter/nfnetlink.h#L56
const NFNL_SUBSYS_IPSET = 6

// IPSET_PROTOCOL is the protocol version.
// http://git.netfilter.org/ipset/tree/include/libipset/linux_ip_set.h
const IPSET_PROTOCOL = 6

// IPSET_MAXNAMELEN is the max length of strings including NUL: set and type identifiers.
const IPSET_MAXNAMELEN = 32

// Message types and commands.
const (
	IPSET_CMD_CREATE  = 2
	IPSET_CMD_DESTROY = 3
	IPSET_CMD_FLUSH   = 4
	IPSET_CMD_ADD     = 9
	IPSET_CMD_DEL     = 10
)

// Attributes at command level.
const (
	IPSET_ATTR_PROTOCOL = 1 /* 1: Protocol version */
	IPSET_ATTR_SETNAME  = 2 /* 2: Name of the set */
	IPSET_ATTR_TYPENAME = 3 /* 3: Typename */
	IPSET_ATTR_REVISION = 4 /* 4: Settype revision */
	IPSET_ATTR_FAMILY   = 5 /* 5: Settype family */
	IPSET_ATTR_DATA     = 7 /* 7: Nested attributes */
)

// CADT specific attributes.
const (
	IPSET_ATTR_IP      = 1
	IPSET_ATTR_CIDR    = 3
	IPSET_ATTR_TIMEOUT = 6
)

// IP specific attributes.
const (
	IPSET_ATTR_IPADDR_IPV4 = 1
	IPSET_ATTR_IPADDR_IPV6 = 2
)

// ATTR flags.
const (
	NLA_F_NESTED        = (1 << 15)
	NLA_F_NET_BYTEORDER = (1 << 14)
)

var nextSeqNr uint32
var nativeEndian binary.ByteOrder

// NetLink struct.
type NetLink struct {
	fd  int
	lsa syscall.SockaddrNetlink
}

// Netlink Options
type Options struct {
	IPv6    bool
	Timeout uint32
}

// Netlink Option func parameter
type Option func(opts *Options)

// New returns a new netlink socket.
func New() (*NetLink, error) {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_NETFILTER)
	if err != nil {
		return nil, err
	}
	// defer syscall.Close(fd)

	nl := &NetLink{fd: fd}
	nl.lsa.Family = syscall.AF_NETLINK

	err = syscall.Bind(fd, &nl.lsa)
	if err != nil {
		syscall.Close(fd)
		return nil, err
	}

	return nl, nil
}

// CreateSet create a ipset.
func (nl *NetLink) CreateSet(setName string, opts ...Option) error {
	if setName == "" {
		return errors.New("ipset: setName must be specified")
	}

	if len(setName) > IPSET_MAXNAMELEN {
		return errors.New("ipset: name too long")
	}

	option := Options{}
	for _, opt := range opts {
		opt(&option)
	}

	var family uint8 = syscall.AF_INET
	if option.IPv6 {
		family = syscall.AF_INET6
	}

	req := NewNetlinkRequest(IPSET_CMD_CREATE|(NFNL_SUBSYS_IPSET<<8), syscall.NLM_F_REQUEST)
	req.AddData(NewNfGenMsg(syscall.AF_INET, 0, 0))
	req.AddData(NewRtAttr(IPSET_ATTR_PROTOCOL, Uint8Attr(IPSET_PROTOCOL)))
	req.AddData(NewRtAttr(IPSET_ATTR_SETNAME, ZeroTerminated(setName)))
	req.AddData(NewRtAttr(IPSET_ATTR_TYPENAME, ZeroTerminated("hash:net")))
	req.AddData(NewRtAttr(IPSET_ATTR_REVISION, Uint8Attr(1)))
	req.AddData(NewRtAttr(IPSET_ATTR_FAMILY, Uint8Attr(family)))

	attrData := NewRtAttr(IPSET_ATTR_DATA|NLA_F_NESTED, nil)
	if option.Timeout != 0 {
		attrData.AddChild(&Uint32Attribute{Type: IPSET_ATTR_TIMEOUT | NLA_F_NET_BYTEORDER, Value: option.Timeout})
	}
	req.AddData(attrData)

	return syscall.Sendto(nl.fd, req.Serialize(), 0, &nl.lsa)
}

// DestroySet destroys a ipset.
func (nl *NetLink) DestroySet(setName string) error {
	if setName == "" {
		return errors.New("setName must be specified")
	}

	req := NewNetlinkRequest(IPSET_CMD_DESTROY|(NFNL_SUBSYS_IPSET<<8), syscall.NLM_F_REQUEST)
	req.AddData(NewNfGenMsg(syscall.AF_INET, 0, 0))
	req.AddData(NewRtAttr(IPSET_ATTR_PROTOCOL, Uint8Attr(IPSET_PROTOCOL)))
	req.AddData(NewRtAttr(IPSET_ATTR_SETNAME, ZeroTerminated(setName)))

	return syscall.Sendto(nl.fd, req.Serialize(), 0, &nl.lsa)
}

// FlushSet flush a ipset.
func (nl *NetLink) FlushSet(setName string) error {
	if setName == "" {
		return errors.New("setName must be specified")
	}

	req := NewNetlinkRequest(IPSET_CMD_FLUSH|(NFNL_SUBSYS_IPSET<<8), syscall.NLM_F_REQUEST)
	req.AddData(NewNfGenMsg(syscall.AF_INET, 0, 0))
	req.AddData(NewRtAttr(IPSET_ATTR_PROTOCOL, Uint8Attr(IPSET_PROTOCOL)))
	req.AddData(NewRtAttr(IPSET_ATTR_SETNAME, ZeroTerminated(setName)))

	return syscall.Sendto(nl.fd, req.Serialize(), 0, &nl.lsa)
}

func (nl *NetLink) HandleAddr(cmd int, setName string, ip netip.Addr, cidr netip.Prefix, opts ...Option) error {
	if setName == "" {
		return errors.New("setName must be specified")
	}

	if len(setName) > IPSET_MAXNAMELEN {
		return errors.New("setName too long")
	}

	req := NewNetlinkRequest(cmd|(NFNL_SUBSYS_IPSET<<8), syscall.NLM_F_REQUEST)
	req.AddData(NewNfGenMsg(syscall.AF_INET, 0, 0))
	req.AddData(NewRtAttr(IPSET_ATTR_PROTOCOL, Uint8Attr(IPSET_PROTOCOL)))
	req.AddData(NewRtAttr(IPSET_ATTR_SETNAME, ZeroTerminated(setName)))

	attrData := NewRtAttr(IPSET_ATTR_DATA|NLA_F_NESTED, nil)

	option := Options{}
	for _, opt := range opts {
		opt(&option)
	}

	if option.Timeout != 0 {
		attrData.AddChild(&Uint32Attribute{Type: IPSET_ATTR_TIMEOUT | NLA_F_NET_BYTEORDER, Value: option.Timeout})
	}

	attrIP := NewRtAttrChild(attrData, IPSET_ATTR_IP|NLA_F_NESTED, nil)

	if ip.Is4() {
		NewRtAttrChild(attrIP, IPSET_ATTR_IPADDR_IPV4|NLA_F_NET_BYTEORDER, ip.AsSlice())
	} else {
		NewRtAttrChild(attrIP, IPSET_ATTR_IPADDR_IPV6|NLA_F_NET_BYTEORDER, ip.AsSlice())
	}

	// for cidr prefix
	if cidr.IsValid() {
		NewRtAttrChild(attrData, IPSET_ATTR_CIDR, Uint8Attr(uint8(cidr.Bits())))
	}

	NewRtAttrChild(attrData, 9|NLA_F_NET_BYTEORDER, Uint32Attr(0))
	req.AddData(attrData)

	return syscall.Sendto(nl.fd, req.Serialize(), 0, &nl.lsa)
}

// NativeEndian get native endianness for the system
func NativeEndian() binary.ByteOrder {
	if nativeEndian == nil {
		var x uint32 = 0x01020304
		if *(*byte)(unsafe.Pointer(&x)) == 0x01 {
			nativeEndian = binary.BigEndian
		} else {
			nativeEndian = binary.LittleEndian
		}
	}
	return nativeEndian
}

func rtaAlignOf(attrlen int) int {
	return (attrlen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}

// NetlinkRequestData interface.
type NetlinkRequestData interface {
	Len() int
	Serialize() []byte
}

// NfGenMsg struct.
type NfGenMsg struct {
	nfgenFamily uint8
	version     uint8
	resID       uint16
}

// NewNfGenMsg returns a new NfGenMsg.
func NewNfGenMsg(nfgenFamily, version, resID int) *NfGenMsg {
	return &NfGenMsg{
		nfgenFamily: uint8(nfgenFamily),
		version:     uint8(version),
		resID:       uint16(resID),
	}
}

// Len returns the length.
func (m *NfGenMsg) Len() int {
	return rtaAlignOf(4)
}

// Serialize serializes NfGenMsg to bytes.
func (m *NfGenMsg) Serialize() []byte {
	native := NativeEndian()

	length := m.Len()
	buf := make([]byte, rtaAlignOf(length))
	buf[0] = m.nfgenFamily
	buf[1] = m.version
	native.PutUint16(buf[2:4], m.resID)
	return buf
}

// RtAttr Extend RtAttr to handle data and children.
type RtAttr struct {
	syscall.RtAttr
	Data     []byte
	children []NetlinkRequestData
}

// NewRtAttr Create a new Extended RtAttr object.
func NewRtAttr(attrType int, data []byte) *RtAttr {
	return &RtAttr{
		RtAttr: syscall.RtAttr{
			Type: uint16(attrType),
		},
		children: []NetlinkRequestData{},
		Data:     data,
	}
}

// NewRtAttrChild Create a new RtAttr obj anc add it as a child of an existing object.
func NewRtAttrChild(parent *RtAttr, attrType int, data []byte) *RtAttr {
	attr := NewRtAttr(attrType, data)
	parent.children = append(parent.children, attr)
	return attr
}

// AddChild adds an existing NetlinkRequestData as a child.
func (a *RtAttr) AddChild(attr NetlinkRequestData) {
	a.children = append(a.children, attr)
}

// Len returns the length of RtAttr.
func (a *RtAttr) Len() int {
	if len(a.children) == 0 {
		return (syscall.SizeofRtAttr + len(a.Data))
	}

	l := 0
	for _, child := range a.children {
		l += rtaAlignOf(child.Len())
	}
	l += syscall.SizeofRtAttr
	return rtaAlignOf(l + len(a.Data))
}

// Serialize the RtAttr into a byte array.
// This can't just unsafe.cast because it must iterate through children.
func (a *RtAttr) Serialize() []byte {
	native := NativeEndian()

	length := a.Len()
	buf := make([]byte, rtaAlignOf(length))

	next := 4
	if a.Data != nil {
		copy(buf[next:], a.Data)
		next += rtaAlignOf(len(a.Data))
	}
	if len(a.children) > 0 {
		for _, child := range a.children {
			childBuf := child.Serialize()
			copy(buf[next:], childBuf)
			next += rtaAlignOf(len(childBuf))
		}
	}

	if l := uint16(length); l != 0 {
		native.PutUint16(buf[0:2], l)
	}
	native.PutUint16(buf[2:4], a.Type)
	return buf
}

// NetlinkRequest is a netlink request.
type NetlinkRequest struct {
	syscall.NlMsghdr
	Data    []NetlinkRequestData
	RawData []byte
}

// NewNetlinkRequest create a new netlink request from proto and flags
// Note the Len value will be inaccurate once data is added until
// the message is serialized.
func NewNetlinkRequest(proto, flags int) *NetlinkRequest {
	return &NetlinkRequest{
		NlMsghdr: syscall.NlMsghdr{
			Len:   uint32(syscall.SizeofNlMsghdr),
			Type:  uint16(proto),
			Flags: syscall.NLM_F_REQUEST | uint16(flags),
			Seq:   atomic.AddUint32(&nextSeqNr, 1),
			// Pid:   uint32(os.Getpid()),
		},
	}
}

// Serialize the Netlink Request into a byte array.
func (req *NetlinkRequest) Serialize() []byte {
	length := syscall.SizeofNlMsghdr
	dataBytes := make([][]byte, len(req.Data))
	for i, data := range req.Data {
		dataBytes[i] = data.Serialize()
		length = length + len(dataBytes[i])
	}
	length += len(req.RawData)

	req.Len = uint32(length)
	b := make([]byte, length)
	hdr := (*(*[syscall.SizeofNlMsghdr]byte)(unsafe.Pointer(req)))[:]
	next := syscall.SizeofNlMsghdr
	copy(b[0:next], hdr)
	for _, data := range dataBytes {
		for _, dataByte := range data {
			b[next] = dataByte
			next = next + 1
		}
	}
	// Add the raw data if any
	if len(req.RawData) > 0 {
		copy(b[next:length], req.RawData)
	}
	return b
}

// AddData add data to request.
func (req *NetlinkRequest) AddData(data NetlinkRequestData) {
	if data != nil {
		req.Data = append(req.Data, data)
	}
}

// AddRawData adds raw bytes to the end of the NetlinkRequest object during serialization.
func (req *NetlinkRequest) AddRawData(data []byte) {
	if data != nil {
		req.RawData = append(req.RawData, data...)
	}
}

// Uint8Attr .
func Uint8Attr(v uint8) []byte {
	return []byte{byte(v)}
}

// Uint16Attr .
func Uint16Attr(v uint16) []byte {
	native := NativeEndian()
	bytes := make([]byte, 2)
	native.PutUint16(bytes, v)
	return bytes
}

// Uint32Attr .
func Uint32Attr(v uint32) []byte {
	native := NativeEndian()
	bytes := make([]byte, 4)
	native.PutUint32(bytes, v)
	return bytes
}

// Uint32Attribute .
type Uint32Attribute struct {
	Type  uint16
	Value uint32
}

// Serialize .
func (a *Uint32Attribute) Serialize() []byte {
	native := NativeEndian()
	buf := make([]byte, rtaAlignOf(8))
	native.PutUint16(buf[0:2], 8)
	native.PutUint16(buf[2:4], a.Type)

	if a.Type&NLA_F_NET_BYTEORDER != 0 {
		binary.BigEndian.PutUint32(buf[4:], a.Value)
	} else {
		native.PutUint32(buf[4:], a.Value)
	}
	return buf
}

// Len .
func (a *Uint32Attribute) Len() int {
	return 8
}

// ZeroTerminated .
func ZeroTerminated(s string) []byte {
	bytes := make([]byte, len(s)+1)
	for i := 0; i < len(s); i++ {
		bytes[i] = s[i]
	}
	bytes[len(s)] = 0
	return bytes
}
