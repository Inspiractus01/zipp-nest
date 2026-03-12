package main

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultNestPort = 9090

// encodeNestCode encodes an IPv4 address into an 8-char hex code XXXX-XXXX.
// Example: "100.86.253.68" → "6456-fd44"
//
//	"192.168.1.5"   → "c0a8-0105"
func encodeNestCode(ip string) (string, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP")
	}
	var val uint32
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return "", fmt.Errorf("invalid IP")
		}
		val = val*256 + uint32(n)
	}
	s := fmt.Sprintf("%08x", val)
	return s[:4] + "-" + s[4:], nil
}
