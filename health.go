package forward

import (
	"log"
	"sync/atomic"

	"github.com/miekg/dns"
)

// For HC we send to . IN NS +norec message to the upstream. Dial timeouts and empty
// replies are considered fails, basically anything else constitutes a healthy upstream.

func (h *host) Check() {
	h.Lock()

	if h.checking {
		h.Unlock()
		return
	}

	h.checking = true
	h.Unlock()

	err := h.send()
	if err != nil {
		log.Printf("[INFO] healtheck of %s failed with %s", h.addr, err)

		HealthcheckFailureCount.WithLabelValues(h.addr).Add(1)

		atomic.AddUint32(&h.fails, 1)
	} else {
		atomic.StoreUint32(&h.fails, 0)
	}

	h.Lock()
	h.checking = false
	h.Unlock()

	return
}

func (h *host) send() error {
	hcping := new(dns.Msg)
	hcping.SetQuestion(".", dns.TypeNS)
	// RecursionDesired is set to true because 9.9.9.9 replies with REFUSED and no
	// question section, which triggers an error.
	hcping.RecursionDesired = true

	_, _, err := h.client.Exchange(hcping, h.addr)
	return err
}

func (h *host) down(maxfails uint32) bool {
	if maxfails == 0 {
		return false
	}

	fails := atomic.LoadUint32(&h.fails)
	return fails > maxfails
}
