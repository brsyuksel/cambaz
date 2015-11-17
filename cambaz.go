package cambaz

import (
	"container/ring"
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"
	"syscall"
	"gopkg.in/fatih/pool.v2"
)

type Cambaz struct {
	bind, forward    *net.TCPAddr
	alternates       []*net.TCPAddr
	poolMin, poolMax int
	verbose          bool

	listener *net.TCPListener

	fwdPool     pool.Pool
	altPoolRing *ring.Ring
}

func (c *Cambaz) Run() error {
	var err error

	printlog(c.verbose, c.bind.String(), "-->", c.forward.String())

	c.listener, err = net.ListenTCP("tcp", c.bind)
	if err != nil {
		return err
	}

	go createPools(c)

	for {
		conn, err := c.listener.AcceptTCP()
		if err != nil {
			printlog(c.verbose, err)

			if err == syscall.EINVAL {
				break
			}
			continue
		}

		go forwardConnection(c, conn)
	}

	return nil
}

func NewCambaz(bind, forward string, alternates mStrSlice, poolSize []int) (*Cambaz, error) {
	c := &Cambaz{
		alternates: []*net.TCPAddr{},
	}
	
	if len(poolSize) == 2 {
		c.poolMin = poolSize[0]
		c.poolMax = poolSize[1]
	}

	if bAddr, err := net.ResolveTCPAddr("tcp", bind); err != nil {
		return nil, err
	} else {
		c.bind = bAddr
	}

	if fAddr, err := net.ResolveTCPAddr("tcp", forward); err != nil {
		return nil, err
	} else {
		c.forward = fAddr
	}

	for _, alternate := range alternates {
		if aAddr, err := net.ResolveTCPAddr("tcp", alternate); err != nil {
			return nil, err
		} else {
			c.alternates = append(c.alternates, aAddr)
		}
	}

	if len(c.alternates) > 0 {
		c.altPoolRing = ring.New(1)
	}
	
	return c, nil
}

func Main() error {
	flag.Parse()
	if *bind == "" || *forward == "" {
		return errors.New("bind and forward parameters must be supplied.")
	}

	pSize := make([]int, 0)
	if *poolSize != "" {
		pStr := strings.Split(*poolSize, ",")
		for _, v := range pStr {
			vInt, err := strconv.Atoi(v)
			if err != nil {
				continue
			}

			if vInt < 0 {
				continue
			}
			pSize = append(pSize, vInt)
		}
	}
	
	if l := len(pSize); l != 0 && l != 2 {
		printlog(true, "Invalid pool size values. Discarded.")
	}

	c, err := NewCambaz(*bind, *forward, alternates, pSize)
	if err != nil {
		return err
	}
	c.verbose = true

	return c.Run()
}
