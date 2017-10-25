package forward

import "sync"

type host struct {
	addr string

	fails uint32
	sync.RWMutex
	checking bool
}

// newHost returns a new host, the fails are set to maxfails+1, i.e.
// the first healthcheck must succeed before we use this host.
func newHost(addr string, maxfails uint32) *host {
	return &host{addr: addr, fails: maxfails + 1}
}

func (h *host) String() string { return h.addr }
