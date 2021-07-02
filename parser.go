package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
	//"github.com/google/gopacket/tcpassembly/tcpreader"
)

func parseReq(r io.Reader, out chan string) int {
	br := bufio.NewReader(r)
	defer io.Copy(io.Discard, br)
	cnt := 0
	for {
		req, err := http.ReadRequest(br)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			log.Printf("ReadRequest fail: %v\n", err)
			continue
		}
		buf := &strings.Builder{}
		req.Write(buf)
		req.Body.Close()
		out <- buf.String()
		cnt++
	}
	return cnt
}

func parseRsp(r io.Reader, out chan string) int {
	br := bufio.NewReader(r)
	defer io.Copy(io.Discard, br)
	cnt := 0
	for {
		rsp, err := http.ReadResponse(br, nil)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			log.Printf("ReadResponse fail: %v\n", err)
			continue
		}
		buf := &strings.Builder{}
		rsp.Write(buf)
		rsp.Body.Close()
		out <- buf.String()
		cnt++
	}
	return cnt
}
