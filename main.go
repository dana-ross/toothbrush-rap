package main

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"log"
)

var upstreamHosts = map[string]string {
	"localhost:8080": "csixty4.com",
}

func RedirectAwareReverseProxy(protocol string, upstreamHosts map[string]string) *httputil.ReverseProxy {

	director := func(req *http.Request) {
			req.URL.Host = upstreamHosts[req.Host]
			req.URL.Scheme = "https"
			fmt.Printf("%+v\n", req)
	}

	modifyResponse := func(r *http.Response) error {
		fmt.Printf("%+v\n", r)
		return nil
	}
	
	return &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: modifyResponse,
	}

}

func main() {
	log.Fatal(http.ListenAndServe(":8080", RedirectAwareReverseProxy("http", upstreamHosts)))
}