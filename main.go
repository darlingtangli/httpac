// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// This binary provides sample code for using the gopacket TCP assembler and TCP
// stream reader.  It reads packets off the wire and reconstructs HTTP requests
// it sees, logging them.
package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

var ip string

var iface = flag.String("i", "lo", "Interface to get packets from")
var fname = flag.String("r", "", "Filename to read from, overrides -i")
var snaplen = flag.Int("s", 16000, "SnapLen for pcap packet capture")
var port = flag.String("p", "11098", "Port to capture")
var output = flag.String("w", "", "Output file to captured http packets, default stderr")

var logAllPackets = flag.Bool("v", false, "Logs every packet in great detail")
var help = flag.Bool("h", false, "Help")

// Build a simple HTTP request parser using tcpassembly.StreamFactory and tcpassembly.Stream interfaces

type endpoint struct {
	host string
	port string
}

func (e endpoint) String() string {
	return e.host + ":" + e.port
}

type endpointPair struct {
	local  endpoint
	remote endpoint
}

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct{}

// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go hstream.run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}

func (h *httpStream) run() {
	if h.net.Src().String() == ip && strings.Contains(*port, h.transport.Src().String()) {
		//if strings.Contains(*port, h.transport.Src().String()) {
		ep := endpointPair{
			local:  endpoint{h.net.Src().String(), h.transport.Src().String()},
			remote: endpoint{h.net.Dst().String(), h.transport.Dst().String()},
		}
		n := parseRsp(&h.r, theDispatcher.responses(ep))
		log.Printf("capture %d http responses from  %s to %s\n", n, ep.remote, ep.local)
	} else {
		ep := endpointPair{
			local:  endpoint{h.net.Dst().String(), h.transport.Dst().String()},
			remote: endpoint{h.net.Src().String(), h.transport.Src().String()},
		}
		n := parseReq(&h.r, theDispatcher.requests(ep))
		log.Printf("capture %d http requests from  %s to %s\n", n, ep.local, ep.remote)
	}
}

func findIp() {
	// 得到所有的(网络)设备
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	// 打印设备信息
	for _, device := range devices {
		if device.Name != *iface {
			continue
		}
		ip = device.Addresses[0].IP.String()
		return
	}
}

func main() {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	if *help {
		flag.Usage()
		return
	}

	//filter := "tcp and host 34.96.111.110 and ("
	filter := "tcp and ("
	ports := strings.Split(*port, ",")
	first := true
	for _, p := range ports {
		if len(p) == 0 {
			continue
		}
		if first {
			filter += "port " + p
			first = false
		} else {
			filter += " or port " + p
		}
	}
	filter += ")"

	// Set up pcap packet capture
	if *fname != "" {
		log.Printf("Reading from pcap dump %q", *fname)
		handle, err = pcap.OpenOffline(*fname)
	} else {
		log.Printf("Starting capture on interface %q, filter %s.", *iface, filter)
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), true, pcap.BlockForever)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	// Set up assembly
	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	//findIp()
	if len(*output) > 0 {
		theDispatcher.o, _ = os.Create(*output)
	}
	go theDispatcher.dump()

	log.Println("reading in packets")
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}
			if *logAllPackets {
				log.Println(packet)
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				//log.Println("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))

		}
	}
}
