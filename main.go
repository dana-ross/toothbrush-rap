package main

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"log"
	"flag"
	"os"
)

var upstreamHosts = map[string]string {
	"localhost:8080": "csixty4.com",
	"localhost:8081": "csixty4.com",
}

func RedirectAwareReverseProxy(protocol string, upstreamHosts map[string]string) *httputil.ReverseProxy {

	director := func(req *http.Request) {
			req.URL.Host = upstreamHosts[req.Host]
			req.Host = upstreamHosts[req.Host]
			req.URL.Scheme = "https"
			fmt.Printf("%+v\n", req)
	}

	modifyResponse := func(r *http.Response) error {
		nextURL := r.Request.URL.String()
		var i int
		for i < 100 {
			client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			} }

			r, err := client.Get(nextURL)

			if err != nil {
				fmt.Println(err)
			}

			fmt.Println("StatusCode:", r.StatusCode)
			fmt.Println(r.Request.URL)

			if r.StatusCode == 200 {
				fmt.Println("Done!")
				break
			} else {
				nextURL = r.Header.Get("Location")
				i += 1
			}
		}
		fmt.Printf("RESPONSE\n%+v\n", r)
		return nil
	}
	
	return &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: modifyResponse,
	}

}

func main() {
	versionParam := flag.Bool("version", false, "display the version number and exit")
	flag.Parse()

	// Handle the --version parameter
	if *versionParam {
		fmt.Printf("%s\n", "0.0.1")
		os.Exit(0)
	}

	log.Println("Starting Toothbrush RAP 0.0.1")
	
	go func(upstreamHosts map[string]string) {
		log.Println("Starting HTTP Listener")
		log.Fatal(http.ListenAndServe(":8080", RedirectAwareReverseProxy("http", upstreamHosts)))
	}(upstreamHosts)

	log.Println("Starting HTTPS Listener")
	log.Fatal(http.ListenAndServe(":8081", RedirectAwareReverseProxy("https", upstreamHosts)))

}