package server

import (
	"fmt"
	"strconv"
	"strings"
)

// parseAddress splits a Go net.Listen-style address (e.g. ":8443",
// "0.0.0.0:8443") into host string and port int.
func parseAddress(addr string) (string, int, error) {
	if addr == "" {
		return "", 0, nil
	}
	// Handle bare port like ":8443"
	if strings.HasPrefix(addr, ":") {
		port, err := strconv.Atoi(addr[1:])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port in %q: %w", addr, err)
		}
		return "", port, nil
	}
	// Split host:port
	lastColon := strings.LastIndex(addr, ":")
	if lastColon < 0 {
		return "", 0, fmt.Errorf("no port in address %q", addr)
	}
	host := addr[:lastColon]
	port, err := strconv.Atoi(addr[lastColon+1:])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in %q: %w", addr, err)
	}
	return host, port, nil
}
