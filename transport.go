package forward

// Copied from coredns/core/dnsserver/address.go

import (
	"strings"
)

// transport returns the protocol of the string s. The second string returns s
// with the prefix chopped off.
func transport(s string) (int, string) {
	switch {
	case strings.HasPrefix(s, _tls+"://"):
		return TransportTLS, s[len(_tls)+3:]
	case strings.HasPrefix(s, _dns+"://"):
		return TransportDNS, s[len(_dns)+3:]
	case strings.HasPrefix(s, _grpc+"://"):
		return TransportGRPC, s[len(_grpc)+3:]
	}
	return TransportDNS, s
}

// Supported transports.
const (
	TransportDNS = iota + 1
	TransportGRPC
	TransportTLS
)

const (
	_dns  = "dns"
	_grpc = "grpc"
	_tls  = "tls"
)
