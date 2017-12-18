package forward

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type host struct {
	addr   string
	client *dns.Client

	transport int
	tlsConfig *tls.Config

	fails uint32
	sync.RWMutex
	checking bool
}

// newHost returns a new host, the fails are set to 1, i.e.
// the first healthcheck must succeed before we use this host.
func newHost(addr string, trans int) *host {
	return &host{addr: addr, fails: 1, transport: trans}
}

// SetTLSConfig sets the TLS config for host h.
func (h *host) SetTLSConfig(cfg *tls.Config) { h.tlsConfig = cfg }

func (h *host) String() string { return h.addr }

// setClient sets and configures the dns.Client in host.
func (h *host) SetClient() {
	c := new(dns.Client)
	c.Net = "tcp"
	c.ReadTimeout = 2 * time.Second
	c.WriteTimeout = 2 * time.Second

	switch h.transport {
	case TransportTLS:
		c.Net = "tcp-tls"
		c.TLSConfig = h.tlsConfig
	}

	h.client = c
}
