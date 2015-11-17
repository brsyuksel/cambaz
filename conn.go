package cambaz

import (
	"gopkg.in/fatih/pool.v2"
	"io"
	"net"
)

const buffByteSize = 256

func pipeConnection(from, to net.Conn, throttle chan bool) {
	buff := make([]byte, buffByteSize)
	for {
		n, err := from.Read(buff)
		if err != nil {
			if err != io.EOF {
				throttle <- true
				break
			}
			throttle <- true
		}

		_, err = to.Write(buff[:n])
		if err != nil {
			throttle <- true
			break
		}
	}
}

func forwardConnection(c *Cambaz, conn net.Conn) {
	defer conn.Close()
	fconn, err := c.fwdPool.Get()
	if err != nil {
		printlog(c.verbose, err)
		
		for i, r := 0, c.altPoolRing; i != r.Len() && r.Next() != nil; i, r = i+1, r.Next() {
			fconn, err = r.Value.(pool.Pool).Get()
			if err != nil {
				printlog(c.verbose, err)
				if i == r.Len()-1 {
					return
				}
				continue
			}
			break
		}
		
	}
	defer fconn.Close()

	throttle := make(chan bool, 2)
	go pipeConnection(conn, fconn, throttle)
	go pipeConnection(fconn, conn, throttle)

	<-throttle
	<-throttle
}
