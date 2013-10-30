// server.go
package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	BufferSize = 1460
)

var (
	conns map[string]net.Conn = make(map[string]net.Conn)
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		req, err := http.ReadRequest(bufio.NewReader(r.Body))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("https", req.URL.Host)
		remote, err := net.Dial("tcp", req.URL.Host)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		log.Println("https", req.URL.Host, "done!")
		rn := rand.New(rand.NewSource(time.Now().UnixNano()))
		s := strconv.FormatInt(rn.Int63(), 10)
		conns[s] = remote

		w.Write([]byte(s))
	})

	http.HandleFunc("/poll", func(w http.ResponseWriter, r *http.Request) {
		rbuf := make([]byte, 0, BufferSize)

		defer r.Body.Close()

		token := r.FormValue("token")
		remote, ok := conns[token]
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		log.Println("/poll: start read from remote")
		n, err := remote.Read(rbuf)
		log.Println("/poll: read from remote", n)
		if err != nil {
			log.Println(n, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			delete(conns, token)
			remote.Close()
			return
		}

		//w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(rbuf)
	})

	http.HandleFunc("/https", func(w http.ResponseWriter, r *http.Request) {
		//log.Println(r.RequestURI, r.URL, r.FormValue("token"))
		defer r.Body.Close()

		token := r.FormValue("token")
		remote, ok := conns[token]
		if !ok {
			log.Println("can't find connect:", token)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		//log.Println(buf.String())
		n, err := io.Copy(remote, r.Body)
		if err != nil {
			log.Println(n, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			delete(conns, token)
			remote.Close()
			return
		}
		log.Println(n, err)
	})

	http.HandleFunc("/http", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		proxy, err := net.Dial("tcp", ":8118")
		if err != nil {
			log.Fatal(err)
		}
		defer proxy.Close()
		defer r.Body.Close()

		i, err := io.Copy(buf, r.Body)
		if err != nil {
			log.Println(i, err)
			return
		}

		//log.Println(buf.String())

		proxy.Write(buf.Bytes())

		//w.Header().Set("Content-Type", "application/octet-stream")

		buf.Reset()
		io.Copy(buf, proxy)
		//log.Println(buf)
		n, err := w.Write(buf.Bytes())
		if err != nil {
			log.Println(n, err)
		}
		//n, _ := io.Copy(w, proxy)
		log.Printf("write back %d data\n", n)
	})

	log.Println("listen on localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
