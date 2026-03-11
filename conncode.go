package main

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultNestPort = 9090

// encodeNestCode encodes a Tailscale IP (100.x.x.x) into a short XXX-XXXX code.
// Example: "100.86.253.68" → "570-0932"
func encodeNestCode(ip string) (string, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP")
	}
	octets := make([]int, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return "", fmt.Errorf("invalid IP")
		}
		octets[i] = n
	}
	val := octets[1]*65536 + octets[2]*256 + octets[3]
	s := fmt.Sprintf("%07d", val)
	return s[:3] + "-" + s[3:], nil
}
