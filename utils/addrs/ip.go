package addrs

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"
)

// SortedIPs type is used to implement sort.Interface for slice of IPNet
type SortedIPs []*net.IPNet

// Returns length of slice
// Implements sort.Interface
func (arr SortedIPs) Len() int {
	return len(arr)
}

// Swap swaps two items in slice identified by indexes
// Implements sort.Interface
func (arr SortedIPs) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// Less returns true if the item in slice at index i in slice
// should be sorted before the element with index j
// Implements sort.Interface
func (arr SortedIPs) Less(i, j int) bool {
	return lessAdrr(arr[i], arr[j])
}

func eqAddr(a *net.IPNet, b *net.IPNet) bool {
	return bytes.Equal(a.IP, b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func lessAdrr(a *net.IPNet, b *net.IPNet) bool {
	if bytes.Equal(a.IP, b.IP) {
		return bytes.Compare(a.Mask, b.Mask) < 0
	}
	return bytes.Compare(a.IP, b.IP) < 0

}

// DiffAddr calculates the difference between two slices of AddrWithPrefix configuration.
// Returns a list of addresses that should be deleted and added to the current configuration to match newConfig.
func DiffAddr(newConfig []*net.IPNet, oldConfig []*net.IPNet) (toBeDeleted []*net.IPNet, toBeAdded []*net.IPNet) {
	var add []*net.IPNet
	var del []*net.IPNet
	//sort
	n := SortedIPs(newConfig)
	sort.Sort(&n)
	o := SortedIPs(oldConfig)
	sort.Sort(&o)

	//compare
	i := 0
	j := 0
	for i < len(n) && j < len(o) {
		if eqAddr(n[i], o[j]) {
			i++
			j++
		} else {
			if lessAdrr(n[i], o[j]) {
				add = append(add, n[i])
				i++
			} else {
				del = append(del, o[j])
				j++
			}
		}
	}

	for ; i < len(n); i++ {
		add = append(add, n[i])
	}

	for ; j < len(o); j++ {
		del = append(del, o[j])
	}
	return del, add
}

// Converts ipv4 configuration from protobuf representation to IPNet
func protoAddrToStruct(addrs []string) ([]*net.IPNet, error) {
	var result []*net.IPNet
	for _, addressWithPrefix := range addrs {
		if addressWithPrefix == "" {
			continue
		}
		parsedIPWithPrefix, _, err := ParseIPWithPrefix(addressWithPrefix)
		if err != nil {
			return result, err
		}
		result = append(result, parsedIPWithPrefix)
	}

	return result, nil
}

// ParseIPWithPrefix returns net representation of the IP address and the ipv6 flag (parsed IP are all the same byte length).
// If the prefix is missing, a default one is added to address (/32 for IPv4, /128 for IPv6)
func ParseIPWithPrefix(input string) (*net.IPNet, bool, error) {
	ipv4Prefix := "/32"
	ipv6Prefix := "/128"

	hasPrefix := strings.Contains(input, "/")

	if hasPrefix {
		ipAddressWithPrefix := strings.Split(input, "/")
		if len(ipAddressWithPrefix) != 2 {
			return nil, false, fmt.Errorf("Incorrect ip address and prefix format: %v", input)
		}
		ip, network, err := net.ParseCIDR(input)
		network.IP = ip
		if err != nil {
			return nil, false, err
		}
		ipv6, err := IsIPv6(ip.String())
		if err != nil {
			return nil, false, err
		}
		if ipv6 {
			return network, true, nil
		}
		return network, false, nil
	}
	// Ip prefix was not set
	ipv6, err := IsIPv6(input)
	if err != nil {
		return nil, false, err
	}
	if ipv6 {
		ip, network, err := net.ParseCIDR(input + ipv6Prefix)
		network.IP = ip
		if err != nil {
			return nil, false, err
		}
		return network, true, nil
	}
	ip, network, err := net.ParseCIDR(input + ipv4Prefix)
	if err != nil {
		return nil, false, err
	}
	network.IP = ip
	return network, false, nil
}

// IsIPv6 returns true if provided IP address is IPv6, false otherwise
func IsIPv6(address string) (bool, error) {
	if strings.Contains(address, ":") {
		return true, nil
	} else if strings.Contains(address, ".") {
		return false, nil
	} else {
		return false, fmt.Errorf("Unknown IP version. Address: %v", address)
	}
}
