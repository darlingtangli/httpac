package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	assembledTrip := make(chan *trip, 100)
	go func() {
		for t := range assembledTrip {
			if !strings.Contains(t.req, "/Ad/Get") {
				continue
			}
			if !strings.Contains(t.rsp, "OK") {
				continue
			}
			//if !strings.Contains(t.req, "sspid=1162") {
			//	continue
			//}
			//post(t.req, t.rsp)
			fmt.Fprintf(d.o, "\n\n######### local: %v remote: %v\n", t.ep.local, t.ep.remote)
			fmt.Fprintf(d.o, t.req)
			fmt.Fprintf(d.o, t.rsp)
			//dump2(t)
		}
	}()

	m := make(map[string]string)
	for t := range d.trips {
		k := t.ep.remote.String() + t.ep.local.String()
		if len(t.req) > 0 {
			// 先来req, 再来rsp
			m[k] = t.req
		}
		if len(t.rsp) > 0 {
			if len(m[k]) > 0 {
				assembledTrip <- &trip{t.ep, m[k], t.rsp}
			}
			m[k] = ""
		}
	}
}

func post(req, rsp string) {
	url := "http://202.168.108.109:60005/xx/debug"
	buf := bytes.NewBuffer(nil)
	//gw := gzip.NewWriter(buf)
	gw, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	gw.Write([]byte(req))
	gw.Write([]byte("\n\n\n"))
	gw.Write([]byte(rsp))
	gw.Close()
	http.Post(url, "", buf)
}
