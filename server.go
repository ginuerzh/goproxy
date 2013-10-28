// server.go
package main

import (
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy, err := net.Dial("tcp", ":8118")
		if err != nil {
			log.Fatal(err)
		}
		defer proxy.Close()

		n, _ := io.Copy(proxy, r.Body)
		r.Body.Close()
		if n == 0 {
			w.Write([]byte("no data"))
			return
		}
		w.Header().Set("Content-type", "application/octet-stream")
		n, _ = io.Copy(w, proxy)
		//log.Printf("write %d data\n", n)
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}
