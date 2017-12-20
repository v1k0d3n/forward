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

	tlsConfig *tls.Config

	fails uint32
	sync.RWMutex
	checking bool
}

// newHost returns a new host, the fails are set to 1, i.e.
// the first healthcheck must succeed before we use this host.
func newHost(addr string) *host {
	return &host{addr: addr, fails: 1}
}

// SetTLSConfig sets the TLS config for host h.
func (h *host) SetTLSConfig(cfg *tls.Config) { h.tlsConfig = cfg }

func (h *host) String() string { return h.addr }

// setClient sets and configures the dns.Client in host.
func (h *host) SetClient() {
	c := new(dns.Client)
	c.Net = "udp"
	c.ReadTimeout = 2 * time.Second
	c.WriteTimeout = 2 * time.Second

	if h.tlsConfig != nil {
		c.Net = "tcp-tls"
		c.TLSConfig = h.tlsConfig
	}

	h.client = c
}
