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
	if v, ok := cache.data[req.Requset]; ok {
		cache.mu.Lock()
		req.Response <- v.value
		cache.mu.Unlock()
		log.Println("Cache: no need to bother archive")
		return
	}

	ch := make(chan string)
	errch := make(chan error)

	archInp <- minidns.Request{
		Requset:  req.Requset,
		Response: ch,
		Err:      errch,
	}

	select {
	case ip := <-ch:
		req.Response <- ip
		add(req.Requset, ip)
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
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		select {
		case newReq := <-input:
			go handle(newReq, archInp)
		case <-t.C:
			go reset()
		}
	}
}
