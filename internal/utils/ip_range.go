package utils

import (
	"fmt"
	"iter"
	"net/netip"
	"strings"
)

type IPRange struct {
	// startIP is the first IP of the range (inclusive)
	startIP netip.Addr
	// endIP is the last IP of the range (inclusive)
	endIP netip.Addr
}

func ParseIPRange(v string) (*IPRange, error) {
	parts := strings.Split(v, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%q must be in format START-END", v)
	}

	startIP, err := netip.ParseAddr(parts[0])
	if err != nil {
		return nil, fmt.Errorf("start IP %q is not valid: %w", parts[0], err)
	}

	endIP, err := netip.ParseAddr(parts[1])
	if err != nil {
		return nil, fmt.Errorf("end IP %q is not valid: %w", parts[0], err)
	}

	if !startIP.Less(endIP) {
		return nil, fmt.Errorf("start IP %q must be less than end IP %q", startIP, endIP)
	}

	return &IPRange{startIP: startIP, endIP: endIP}, nil
}

// Iterate returns an iterator over all of the addresses of this IPRange.
//
// Example usage:
//
//	r, _ := utils.ParseIPRange("10.0.0.10-10.0.0.20")
//	for addr := range iprange.Iterate() {
//		fmt.Println(addr.String())  // 10.0.0.10, 10.0.0.11, ..., 10.0.0.20
//	}
func (r *IPRange) Iterate() iter.Seq[string] {
	current := r.startIP
	return func(yield func(string) bool) {
		for !r.endIP.Less(current) {
			if !yield(current.String()) {
				return
			}
			current = current.Next()
		}
	}
}

// IPRanges is one or more (possibly overlapping) range of IPs
type IPRanges []*IPRange

func ParseIPRanges(v string) (IPRanges, error) {
	if v == "" {
		return nil, nil
	}

	parts := strings.Split(v, ",")
	ranges := make(IPRanges, 0, len(parts))
	for _, part := range parts {
		r, err := ParseIPRange(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse range %q: %w", part, err)
		}
		ranges = append(ranges, r)
	}

	return ranges, nil
}

func (r IPRanges) Iterate() iter.Seq[string] {
	return func(yield func(string) bool) {
		// iterate over all IPRanges one by one
		for _, ipRange := range r {
			for ip := range ipRange.Iterate() {
				if !yield(ip) {
					return
				}
			}
		}
	}
}
