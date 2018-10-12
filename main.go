package main

import (
	"sync"
	"net/url"
	"log"
	"time"
)

type concurrentStorage struct {
	sync.Mutex
	domain string
	urls map[url.URL]bool
}

func NewConcurrentStorage(d string) *concurrentStorage{
	return &concurrentStorage{
		domain: d,
		urls: map[url.URL]bool{},
	}
}

func (c *concurrentStorage) Add(u url.URL) (url.URL, bool) {

	//newURL, err := url.Parse(u)
	//if err != nil {
	//	// TODO: log why cant parse `u`
	//	return url.URL{}, false
	//}



	c.Lock()
	defer c.Unlock()
	if _, ok := c.urls[u]; ok{
		return url.URL{}, false
	}
	c.urls[u] = true
	return u, true
}

/*
Patterns:
	if Starts With "/" :
		append prefix "http://{domain}"
	if contains \W{domain} (ex: \Wtarger.com)
		ignore
	if Starts With "http"
*/

func main() {

	ch := make(chan url.URL, 5)
	domain := "target.com"
	targetURL, _:= url.Parse("https://www.target.com/")

	urlSet := NewConcurrentStorage(domain)

	//foo := func(ch chan url.URL) {
	//	for {
	//		select {
	//		case rawURL := <- ch:
	//			if url, ok := urlSet.Add(rawURL); ok{
	//				// TODO: instead of printing, craw the rawURL
	//				fmt.Printf("%s\n",url.String())
	//			}
	//		}
	//	}
	//}
	//
	//go foo(ch)
	//go foo(ch)


	go crawl(*urlSet, ch)
	go crawl(*urlSet, ch)
	go crawl(*urlSet, ch)

	ch <- *targetURL


	time.Sleep(10 * time.Second)

	log.Printf("total in %d seconds = %d", 10, len(urlSet.urls))
	//fmt.Printf("Bye")


}








