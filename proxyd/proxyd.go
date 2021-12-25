package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func main() {
	port := flag.Int("port", 8080, "port")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), proxy))
}
