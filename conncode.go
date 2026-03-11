package main

import (
	"encoding/base64"
	"fmt"
)

// buildConnCode encodes host:port:token into a single base64 string.
func buildConnCode(host string, port int, token string) string {
	raw := fmt.Sprintf("%s:%d:%s", host, port, token)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(raw))
}
