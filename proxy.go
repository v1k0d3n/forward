package forward

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type proxy struct {
	host *host

	transport *transport

	// copied from Forward.
	hcInterval time.Duration
	forceTCP   bool

	stop chan bool

	sync.RWMutex
}

func newProxy(addr string) *proxy {
	host := newHost(addr)

	p := &proxy{
		host:       host,
		hcInterval: hcDuration,
		stop:       make(chan bool),
		transport:  newTransport(host),
	}
	return p
}

// SetTLSConfig sets the TLS config in the lower p.host.
func (p *proxy) SetTLSConfig(cfg *tls.Config) { p.host.tlsConfig = cfg }

// SetExpire sets the expire duration in the lower p.host.
func (p *proxy) SetExpire(expire time.Duration) { p.host.expire = expire }

func (p *proxy) close() { p.stop <- true }

// Connect connect to the host in p with the configured transport.
func (p *proxy) Dial(proto string) (*dns.Conn, error) { return p.transport.Dial(proto) }

// Yield returns the connection to the pool.
func (p *proxy) Yield(c *dns.Conn) { p.transport.Yield(c) }

// Down returns if this proxy is up or down.
func (p *proxy) Down(maxfails uint32) bool { return p.host.down(maxfails) }

func (p *proxy) healthCheck() {

	// stop channel
	p.host.SetClient()

	p.host.Check()
	tick := time.NewTicker(p.hcInterval)
	for {
		select {
		case <-tick.C:
			p.host.Check()
		case <-p.stop:
			return
		}
	}
}

const (
	dialTimeout = 2 * time.Second
	hcDuration  = 500 * time.Millisecond
)
