package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	minidns "github.com/mamad-nik/mini-dns"
)

type entry struct {
	value  string
	ticker *time.Ticker
}

type cacheType struct {
	data map[string]entry
	mu   sync.Mutex
}

var cache cacheType

func add(domain, ip string) {
	e := entry{
		value:  ip,
		ticker: time.NewTicker(10 * time.Second),
	}
	cache.mu.Lock()
	cache.data[domain] = e
	cache.mu.Unlock()
}

func handle(req minidns.Request, archInp chan<- minidns.Request) {
	fmt.Println(cache.data)
	if v, ok := cache.data[req.Domain]; ok {
		cache.mu.Lock()
		req.IP <- v.value
		cache.mu.Unlock()
		log.Println("Cache: no need to bother archive")
		return
	}

	ch := make(chan string)
	errch := make(chan error)

	archInp <- minidns.Request{
		Domain: req.Domain,
		IP:     ch,
		Err:    errch,
	}

	select {
	case ip := <-ch:
		req.IP <- ip
		add(req.Domain, ip)
	case err := <-errch:
		req.Err <- err
	}
}

func reset() {

	for _, v := range cache.data {
		go func(m entry) {
			<-m.ticker.C
			cache.mu.Lock()
			m.ticker.Stop()
			delete(cache.data, m.value)
			cache.mu.Unlock()
		}(v)
	}
}

func Run(input <-chan minidns.Request, archInp chan<- minidns.Request) {
	cache.data = make(map[string]entry)
	go reset()
	for newReq := range input {
		go handle(newReq, archInp)
	}
}
