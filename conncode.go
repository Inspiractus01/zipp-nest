package main

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultNestPort = 9090

// encodeNestCode encodes any IPv4 address into a 10-digit decimal code XXXXX-XXXXX.
// Example: "100.86.253.68" → "16834-22532"
//          "192.168.1.5"   → "32322-35781"
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
	s := fmt.Sprintf("%010d", val)
	return s[:5] + "-" + s[5:], nil
}
