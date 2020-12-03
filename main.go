package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
)

var VERSION = "0.0.2"

var upstreamHosts = map[string]string{
	"localhost:8080": "csixty4.com",
	"localhost:8081": "csixty4.com",
}

func makeURLAbsolute(r *http.Request, protocol string) string {
	urlStr := r.URL.String()
	if u, err := url.Parse(urlStr); err == nil {
		// If url was relative, make absolute by
		// combining with request path.
		// The browser would probably do this for us,
		// but doing it ourselves is more reliable.

		// NOTE(rsc): RFC 2616 says that the Location
		// line must be an absolute URI, like
		// "http://www.google.com/redirect/",
		// not a path like "/redirect/".
		// Unfortunately, we don't know what to
		// put in the host name section to get the
		// client to connect to us again, so we can't
		// know the right absolute URI to send back.
		// Because of this problem, no one pays attention
		// to the RFC; they all send back just a new path.
		// So do we.
		oldpath := r.URL.Path
		if oldpath == "" { // should not happen, but avoid a crash if it does
			oldpath = "/"
		}
		if u.Scheme == "" {
			// no leading http://server
			if urlStr == "" || urlStr[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				urlStr = olddir + urlStr
			}

			var query string
			if i := strings.Index(urlStr, "?"); i != -1 {
				urlStr, query = urlStr[:i], urlStr[i:]
			}

			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(urlStr, "/")
			urlStr = path.Clean(urlStr)
			if trailing && !strings.HasSuffix(urlStr, "/") {
				urlStr += "/"
			}
			urlStr += query

			urlStr = fmt.Sprintf("%s://%s%s", protocol, r.Host, urlStr)
		}
	}

	return urlStr
}

func RedirectAwareReverseProxy(protocol string, upstreamHosts map[string]string) *httputil.ReverseProxy {

	director := func(req *http.Request) {

		req.Header.Set("x-tr-originalurl", makeURLAbsolute(req, protocol))
		req.URL.Host = upstreamHosts[req.Host]
		req.Host = upstreamHosts[req.Host]
		req.URL.Scheme = "https"
	}

	modifyResponse := func(r *http.Response) error {
		originalURL := r.Request.Header.Get("x-tr-originalurl")
		nextURL := r.Request.URL.String()
		var i int
		for i < 100 {
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}}

			r, err := client.Get(nextURL)

			if err != nil {
				fmt.Println(err)
			}

			if r.StatusCode == 200 {
				break
			} else {
				nextURL = r.Header.Get("Location")
				i += 1
			}
		}
		log.Printf("Original: %s Location: %s", originalURL, nextURL)
		return nil
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		Transport:      NewCachedRoundTripper(10),
	}

}

func main() {
	versionParam := flag.Bool("version", false, "display the version number and exit")
	flag.Parse()

	// Handle the --version parameter
	if *versionParam {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	log.Printf("Starting Toothbrush RAP %s\n", VERSION)

	go func(upstreamHosts map[string]string) {
		log.Println("Starting HTTP Listener")
		log.Fatal(http.ListenAndServe(":8080", RedirectAwareReverseProxy("http", upstreamHosts)))
	}(upstreamHosts)

	log.Println("Starting HTTPS Listener")
	log.Fatal(http.ListenAndServe(":8081", RedirectAwareReverseProxy("https", upstreamHosts)))

}
