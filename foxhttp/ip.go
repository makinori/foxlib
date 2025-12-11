package foxhttp

import (
	"net/http"
	"regexp"
	"strings"
)

var (
	// https://www.ditig.com/validating-ipv4-and-ipv6-addresses-with-regexp
	ipv6Regex = regexp.MustCompile(`^((?:[0-9A-Fa-f]{1,4}:){7}[0-9A-Fa-f]{1,4}|(?:[0-9A-Fa-f]{1,4}:){1,7}:|:(?::[0-9A-Fa-f]{1,4}){1,7}|(?:[0-9A-Fa-f]{1,4}:){1,6}:[0-9A-Fa-f]{1,4}|(?:[0-9A-Fa-f]{1,4}:){1,5}(?::[0-9A-Fa-f]{1,4}){1,2}|(?:[0-9A-Fa-f]{1,4}:){1,4}(?::[0-9A-Fa-f]{1,4}){1,3}|(?:[0-9A-Fa-f]{1,4}:){1,3}(?::[0-9A-Fa-f]{1,4}){1,4}|(?:[0-9A-Fa-f]{1,4}:){1,2}(?::[0-9A-Fa-f]{1,4}){1,5}|[0-9A-Fa-f]{1,4}:(?:(?::[0-9A-Fa-f]{1,4}){1,6})|:(?:(?::[0-9A-Fa-f]{1,4}){1,6}))$`)
)

func IsValidIPv6(address string) bool {
	return ipv6Regex.MatchString(address)
}

func GetIPAddress(r *http.Request) string {
	ipAddress := r.Header.Get("X-Real-IP")

	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	}

	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	ipAddress = strings.SplitN(ipAddress, ",", 2)[0]
	ipAddress = strings.TrimSpace(ipAddress)

	if IsValidIPv6(ipAddress) {
		return ipAddress
	}

	ipAddress = portRegexp.ReplaceAllString(ipAddress, "")

	ipAddress = strings.TrimPrefix(ipAddress, "[")
	ipAddress = strings.TrimSuffix(ipAddress, "]")

	return ipAddress
}
