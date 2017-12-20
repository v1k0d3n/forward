package forward

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type proxy struct {
	host *host

	Transport *Transport

	// copied from Forward.
	hcInterval time.Duration
	forceTCP   bool

	closed bool

	sync.RWMutex
}

func newProxy(addr string) *proxy {
	host := newHost(addr)

	p := &proxy{
		host:       host,
		closed:     false,
		hcInterval: hcDuration,
		Transport:  NewTransport(host),
	}
	return p
}

// SetTLSConfig sets the TLS config lower p.host.
func (p *proxy) SetTLSConfig(cfg *tls.Config) { p.host.SetTLSConfig(cfg) }

func (p *proxy) Closed() bool {
	p.RLock()
	b := p.closed
	p.RUnlock()
	return b
}

func (p *proxy) SetClosed(b bool) {
	p.Lock()
	p.closed = b
	p.Unlock()
}

// Connect connect to the host in p with the configured transport.
func (p *proxy) Dial(proto string) (*dns.Conn, error) { return p.Transport.Dial(proto) }

// Yield returns the connection to the pool.
func (p *proxy) Yield(c *dns.Conn) { p.Transport.Yield(c) }

// Down returns if this proxy is up or down.
func (p *proxy) Down(maxfails uint32) bool { return p.host.down(maxfails) }

func (p *proxy) healthCheck() {

	// stop channel
	p.host.SetClient()

	for !p.Closed() {
		p.host.Check()
		time.Sleep(p.hcInterval)
	}
}

const (
	dialTimeout = 2 * time.Second
	hcDuration  = 500 * time.Millisecond
)
