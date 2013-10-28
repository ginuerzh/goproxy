// goproxy project goproxy.go
package main

import (
	"bytes"
	"log"
	"net"
	"net/http"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	var rbuf, wbuf [1460]byte

	defer conn.Close()

	proxy, err := net.Dial("tcp", ":8118")
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Close()

	n, err := conn.Read(rbuf[:])
	if err != nil {
		log.Println(err)
	}
	log.Printf("read %d data\n", n)

	request, _ := http.NewRequest("POST", "106.187.48.51:8080", bytes.NewBuffer(rbuf[:n]))
	request.Header.Add("Content-type", "application/octet-stream")

	if n > 0 {
		var data [4096]byte
		buf := bytes.NewBuffer(data[:])
		request.Write(buf)
		log.Println(string(buf.Bytes()))
		n, err = proxy.Write(buf.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("write %d data\n", n)
	}

	for {
		n, err := proxy.Read(wbuf[:])
		if n > 0 {
			conn.Write(wbuf[:n])
		}
		if err != nil {
			log.Println(err)
			break
		}
	}
}
