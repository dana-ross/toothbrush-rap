package main

import (
	"net/http"
	"log"
	"time"
	"github.com/ReneKroon/ttlcache"
)

type CachedRoundTripper struct {
	cache *ttlcache.Cache
}

func (crt *CachedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	cachedResp, err := crt.cache.Get(r.Header.Get("x-tr-originalurl"))

	if err == nil {
		log.Println("Returning cached response")
		return cachedResp.(*http.Response), nil
	}

	resp, err :=  http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	log.Println("Caching response")
	crt.cache.Set(r.Header.Get("x-tr-originalurl"), resp)
	return resp, nil
}

func NewCachedRoundTripper(ttl int) http.RoundTripper {
	log.Printf("Initializing Cache with %d second TTL", ttl)
	crt := new(CachedRoundTripper)
	crt.cache = ttlcache.NewCache()
	crt.cache.SetTTL(time.Duration(time.Duration(ttl) * time.Second))
	return crt
}