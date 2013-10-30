// goproxy project goproxy.go
package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

var (
	proxy  string
	server string
)

const (
	connectURI = "/connect"
	httpsURI   = "/https"
	httpURI    = "/http"
	pollURI    = "/poll"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&proxy, "P", "", "http proxy for forward")
	flag.StringVar(&server, "s", "localhost:8000", "the server that client connecting to")
	flag.Parse()
}

func main() {
	if !strings.HasPrefix(server, "http://") {
		server = "http://" + server
	}
	log.Println(proxy, server)

	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func request(method string, urlStr string, body io.Reader, proxy net.Conn) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, urlStr, body)

	if proxy != nil {
		if err != nil {
			return
		}
		err = req.WriteProxy(proxy)
		if err != nil {
			return
		}
		data, err := ioutil.ReadAll(proxy)
		if err != nil {
			return nil, err
		}
		resp, err = http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), req)
	} else {
		client := new(http.Client)
		resp, err = client.Do(req)
	}
	return
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var proxyConn net.Conn

	if len(proxy) > 0 {
		pConn, err := net.Dial("tcp", proxy)
		if err != nil {
			log.Println(err)
			return
		}
		defer pConn.Close()
		proxyConn = pConn
	}

	reqData, err := read(conn)
	if err != nil {
		log.Println(reqData.Len(), err)
		return
	}
	//log.Println(reqData)
	r, err := resolve(reqData.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if r.Method == "CONNECT" {
		resp, err := request("POST", server+connectURI, reqData, proxyConn)
		if err != nil {
			log.Println(err)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			return
		}
		token := string(data)
		//log.Println(token)
		buf := new(bytes.Buffer)
		buf.WriteString("HTTP/1.0 200 Connection established\r\nProxy-agent: go-http-tunnel\r\n\r\n")
		conn.Write(buf.Bytes())

		go func() {
			for {
				data, err := read(conn)
				log.Println("read from client", data.Len())
				if err != nil {
					log.Println(err)
					break
				}

				resp, err := request("POST", server+httpsURI+"?token="+token, data, proxyConn)
				if err != nil {
					break
				}
				log.Println(err, resp.StatusCode)
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					break
				}
			}
		}()

		for {
			resp, err := request("GET", server+pollURI+"?token="+token, nil, proxyConn)
			if err != nil {
				log.Println(err)
				break
			}
			defer resp.Body.Close()

			log.Println(err, resp.StatusCode)
			if resp.StatusCode != http.StatusOK {
				break
			}
			_, err = io.Copy(conn, resp.Body)
			if err != nil {
				log.Println(err)
				break
			}
		}

		return
	}

	resp, err := request("POST", server+httpURI, reqData, proxyConn)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.Copy(conn, resp.Body)
}

func read(conn net.Conn) (*bytes.Buffer, error) {
	var rbuf [1460]byte
	buf := new(bytes.Buffer)

	n, err := conn.Read(rbuf[:])
	buf.Write(rbuf[:n])
	//log.Println("read from connect", buf.Len(), err)
	return buf, err
}

func resolve(reqData []byte) (*http.Request, error) {
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(reqData)))
	return request, err
}
