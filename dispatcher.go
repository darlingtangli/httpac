package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type connection struct {
	requests  chan string
	responses chan string
}

type trip struct {
	ep  endpointPair
	req string
	rsp string
}

type dispatcher struct {
	mu    *sync.Mutex
	conns map[endpointPair]*connection
	trips chan *trip
	o     io.Writer
}

var theDispatcher = &dispatcher{
	mu:    &sync.Mutex{},
	conns: make(map[endpointPair]*connection),
	trips: make(chan *trip, 100),
	o:     os.Stderr,
}

func (d *dispatcher) conn(ep endpointPair) *connection {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.conns[ep]
	if !ok {
		c = &connection{
			requests:  make(chan string),
			responses: make(chan string),
		}
		d.conns[ep] = c
		go d.onCreate(ep, c)
	}
	return c
}

func (d *dispatcher) requests(ep endpointPair) chan string {
	return d.conn(ep).requests
}

func (d *dispatcher) responses(ep endpointPair) chan string {
	return d.conn(ep).responses
}

func (d *dispatcher) onCreate(ep endpointPair, c *connection) {
	//	for {
	//		req := <-c.requests
	//		rsp := <-c.responses
	//		d.trips <- &trip{ep, req, rsp}
	//	}
	for {
		select {
		case req := <-c.requests:
			d.trips <- &trip{ep, req, ""}
		case rsp := <-c.responses:
			d.trips <- &trip{ep, "", rsp}
		}
	}
}

func (d *dispatcher) dump() {
	for trip := range d.trips {
		if len(trip.req) > 0 {
			fmt.Fprintf(d.o, "---------------- REQ: %s -> %s --------------------\n", trip.ep.remote, trip.ep.local)
			fmt.Fprintf(d.o, "%s\n", trip.req)
		}
		if len(trip.rsp) > 0 {
			fmt.Fprintf(d.o, "\n---------------- RSP: %s -> %s --------------------\n", trip.ep.local, trip.ep.remote)
			fmt.Fprintf(d.o, "%s\n", trip.rsp)
		}
	}
}
