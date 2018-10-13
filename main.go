package main

//
//func main() {
//	var domain string
//	ch := make(chan url.URL, 5)
//
//	flag.StringVar(&domain, "host", "example.com", "url to scrape")
//
//	fmt.Printf("RECEIVED %s\n", domain)
//
//	targetURL, _:= url.Parse(domain)
//
//	// TODO: write function to find a valid schema by requesting with multiple versions of the url
//	if targetURL.Scheme == "" {
//		targetURL.Scheme = "https"
//	}
//
//	urlSet := newConcurrentStorage(targetURL.Host)
//
//	go crawl(*urlSet, ch)
//	go crawl(*urlSet, ch)
//	go crawl(*urlSet, ch)
//
//	ch <- *targetURL
//
//	time.Sleep(2 * time.Second)
//
//	log.Printf("total in %d seconds = %d", 2, len(urlSet.urls))
//	//for u := range urlSet.urls {
//	//	log.Printf("url: %s\n", u.String())
//	//}
//}
//
//






