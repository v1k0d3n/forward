/*
 * Copyright (c) 2016 Felipe Cavalcanti <fjfcavalcanti@gmail.com>
 * Author: Felipe Cavalcanti <fjfcavalcanti@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

/* General idea kept, but completely rewritten for use in CoreDNS - Miek Gieben 2017 */

package forward

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type proxy struct {
	host *host

	transport int // transport to use dns, tls or grpc (grpc not implemented yet)

	// copied from Forward.
	hcInterval time.Duration
	forceTCP   bool

	closed bool

	sync.RWMutex
}

func newProxy(addr string, transport int) *proxy {
	return &proxy{
		host:       newHost(addr, transport),
		transport:  transport,
		closed:     false,
		hcInterval: hcDuration,
	}
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
func (p *proxy) Connect(proto string) (*dns.Conn, error) {
	switch p.transport {
	case TransportDNS:
		return dns.DialTimeout(proto, p.host.addr, dialTimeout)
	case TransportTLS:
		return dns.DialTimeoutWithTLS("tcp", p.host.addr, p.host.tlsConfig, dialTimeout)
	}

	return dns.DialTimeout(proto, p.host.addr, dialTimeout)
}

// Down returns if this proxy is up or down.
func (p *proxy) Down(maxfails uint32) bool { return p.host.down(maxfails) }

func (p *proxy) healthCheck() {

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
