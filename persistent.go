package forward

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

var collect = 10 * time.Second

type persistConn struct {
	c    *dns.Conn
	used time.Time
}

type connErr struct {
	c   *dns.Conn
	err error
}

//map string connectMethodKey (transport one??)
type Transport struct {
	conns map[string][]*persistConn //  Buckets for udp, tcp and tcp-tls
	host  *host

	dial  chan string
	yield chan connErr
	ret   chan connErr

	stop chan bool
}

func NewTransport(h *host) *Transport {
	t := &Transport{
		conns: make(map[string][]*persistConn),
		host:  h,
		dial:  make(chan string),
		yield: make(chan connErr),
		ret:   make(chan connErr),
		stop:  make(chan bool),
	}
	go t.connManager()
	return t
}

func (t *Transport) connManager() {

Wait:
	for {
		select {
		case proto := <-t.dial:
			// Yes O(n), shouldn't put millions in here.
			i := 0
			for i = 0; i < len(t.conns[proto]); i++ {
				pc := t.conns[proto][i]
				if time.Since(pc.used) < collect {
					t.conns[proto] = t.conns[proto][i+1:]
					t.ret <- connErr{pc.c, nil}
					continue Wait
				}

				pc.c.Close()
			}

			t.conns[proto] = t.conns[proto][i:]

			go func() {
				if proto != "tcp-tls" {
					c, err := dns.DialTimeout(proto, t.host.addr, dialTimeout)
					t.ret <- connErr{c, err}
					return
				}

				c, err := dns.DialTimeoutWithTLS("tcp", t.host.addr, t.host.tlsConfig, dialTimeout)
				t.ret <- connErr{c, err}
			}()

		case conn := <-t.yield:
			// no proto here, infer from config and conn
			if _, ok := conn.c.Conn.(*net.UDPConn); ok {
				t.conns["udp"] = append(t.conns["udp"], &persistConn{conn.c, time.Now()})
				continue Wait
			}

			if t.host.tlsConfig == nil {
				t.conns["tcp"] = append(t.conns["tcp"], &persistConn{conn.c, time.Now()})
				continue Wait
			}

			t.conns["tcp-tls"] = append(t.conns["tcp-tls"], &persistConn{conn.c, time.Now()})

		case <-t.stop:
			return
		}
	}
}

func (t *Transport) Dial(proto string) (*dns.Conn, error) {
	t.dial <- proto
	c := <-t.ret
	return c.c, c.err
}

func (t *Transport) Yield(c *dns.Conn) {
	t.yield <- connErr{c, nil}
}

// Stop stops the transports.
func (t *Transport) Stop() { t.stop <- true }
