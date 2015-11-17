package cambaz

import (
	"container/ring"
	"github.com/cenkalti/backoff"
	"gopkg.in/fatih/pool.v2"
	"net"
	"sync"
	"time"
)

type PoolLike struct {
	addr *net.TCPAddr
	conn net.Conn
}

func (p *PoolLike) Get() (net.Conn, error) {
	var err error
	p.conn, err = net.Dial(p.addr.Network(), p.addr.String())
	return p.conn, err
}

func (p *PoolLike) Close() {
	p.conn.Close()
}

func (p *PoolLike) Len() int {
	return 1
}

func createPools(c *Cambaz) {
	if c.poolMin == 0 {
		c.fwdPool = &PoolLike{
			addr: c.forward,
		}

		for _, altAddr := range c.alternates {
			p := &PoolLike{addr: altAddr}

			if c.altPoolRing.Value == nil {
				c.altPoolRing.Value = p
				continue
			}

			c.altPoolRing.Link(&ring.Ring{Value: p})
			c.altPoolRing = c.altPoolRing.Next()
		}

		return
	}

	ringSync := &sync.Mutex{}

	forceCreatePool := func(addr *net.TCPAddr, isForward bool) {
		factory := func() (net.Conn, error) {
			return net.Dial(addr.Network(), addr.String())
		}

		var p pool.Pool
		poolCreator := func() (err error) {
			p, err = pool.NewChannelPool(c.poolMin, c.poolMax, factory)
			return
		}

		poolLog := func(err error, wait time.Duration) {
			printlog(c.verbose, err)
		}

		err := backoff.RetryNotify(poolCreator, backoff.NewExponentialBackOff(), poolLog)
		if err != nil {
			printlog(c.verbose, err)
			return
		}

		if isForward {
			c.fwdPool = p
			return
		}

		ringSync.Lock()
		if c.altPoolRing.Value == nil {
			c.altPoolRing.Value = p
			ringSync.Unlock()
			return
		}
		c.altPoolRing.Link(&ring.Ring{Value: p})
		c.altPoolRing = c.altPoolRing.Next()
		ringSync.Unlock()
	}

	go forceCreatePool(c.forward, true)
	for _, altAddr := range c.alternates {
		go forceCreatePool(altAddr, false)
	}
}
