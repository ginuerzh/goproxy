// goproxy project goproxy.go
package main

import (
	"bufio"
	"bytes"
	"io"
	//"io/ioutil"
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
	buf := new(bytes.Buffer)

	defer conn.Close()

	proxy, err := net.Dial("tcp", ":8118")
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Close()

	for {
		n, err := conn.Read(rbuf[:])
		//log.Println(n, err)
		if n > 0 {
			buf.Write(rbuf[:n])
		}

		if n < 1460 || err == io.EOF {
			break
		}
	}
	//log.Printf("read %d data from request\n", buf.Len())
	request, _ := http.NewRequest("POST", "http://106.187.48.51:8000/", buf)
	request.WriteProxy(proxy)

	buf.Reset()
	for {
		n, err := proxy.Read(wbuf[:])
		if n > 0 {
			buf.Write(wbuf[:n])
		}
		//log.Println(n, err)
		if err == io.EOF {
			break
		}
	}
	//log.Printf("read %d data from proxy\n", buf.Len())

	resp, err := http.ReadResponse(bufio.NewReader(buf), request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	io.Copy(conn, resp.Body)
}
