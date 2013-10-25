// goproxy project goproxy.go
package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Body)
}

func main() {
	http.HandleFunc("/", handler)
	log.Printf("About to listen on 8080. Go to https://127.0.0.1:8080/")
	err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
