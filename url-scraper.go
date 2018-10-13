/*
	Crawl a host and save all relevant pages to local storage
*/
package main

import (
	"log"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"net/url"

	"bytes"
	"io/ioutil"
	"strings"
	"sync"
	"errors"
	"time"
	"flag"
)

// concurrentStorage acts as a set. A common storage point for multiple go routines and
// as a validator, to avoid processing urls that have already been processed by other routines.
type concurrentStorage struct {
	sync.Mutex
	domain string
	urls map[url.URL]bool
	urlsSize int
}

func newConcurrentStorage(d string) *concurrentStorage{
	return &concurrentStorage{
		domain: d,
		urls: map[url.URL]bool{},
	}
}

// Return true if the URL is unseen and was saved.
//
// add saves a URL iff it hasn't been processed by a go routine. If it
// cannot save it, then returns an empty URL and false to let the caller
// know not to process it.
func (c *concurrentStorage) add(u url.URL) (bool) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.urls[u]; ok{
		return false
	}
	c.urls[u] = true
	c.urlsSize++
	return true
}

func (c *concurrentStorage) size() int {
	c.Lock()
	defer c.Unlock()
	return c.urlsSize
}



// scrape visits a page and extracts all the valid urls for the given domain
// Returns error if the target URL is empty, cannot be scrapped by access over HTTP,
// urls cannot be scraped.
func scrape(u url.URL) ([]url.URL, error) {

	if strings.Trim(u.String(), " ") == ""{
		return []url.URL{}, errors.New("empty url")
	}

	pageReadCloser, err := getHttp(u)
	defer pageReadCloser.Close()
	if err != nil {
		log.Printf("failed to get pageReadCloser at u=%s. err=%s\n", u, err)
		return []url.URL{}, nil
	}

	page, err := ioutil.ReadAll(pageReadCloser)
	if err != nil {
		log.Printf("Could not read page buffer for url=%s\n", u.String())
		return []url.URL{}, err
	}

	savePage(u, page)

	urls, err := getUrls(page, u.Host)
	if err != nil {
		log.Printf("failed to extract valid urls for pageReadCloser at u=%s. err=%s\n", u, err)
		return []url.URL{}, err
	}

	return urls, nil
}

// adds missing pieces to a URL and then validates it.
// if is an invalid/non-accessible URL then return false
func sanitizeUrl(href string, domain string) (url.URL, bool){
	if strings.Trim(href, " ") == ""{
		return url.URL{}, false
	}

	u, err := url.Parse(href)
	if err != nil {
		log.Println(err)
		return url.URL{}, false
	}

	if u.Host == ""{
		u.Host = domain
	} else if u.Host != domain || u.Path == "/" || u.Path == ""{
		return url.URL{}, false
	}

	if u.Scheme == ""{
		u.Scheme = "https"
	}

	// Ignore alien schemas [ mailto, ftp, etc ]
	if !strings.Contains(u.Scheme, "http") {
		return url.URL{}, false
	}

	// TODO: Check URL is accessible

	return *u, true
}

// Extract the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

// Get the contents of a web page
// Return error if the request fails
func getHttp(url url.URL) (io.ReadCloser, error) {
	resp, err := http.Get(url.String())
	if err != nil {
		log.Printf("HTTP failed to GET url=%s. error=%s\n", url.String(), err)
		return nil, err
	}

	return resp.Body, nil
}

// Get only urls of the specified domain given the body of a web page
func getUrls(body []byte, domain string) ([]url.URL, error) {

	// holds only valid urls
	var urls []url.URL

	reader := bytes.NewReader(body)
	tokenizer := html.NewTokenizer(reader)

	infinitefor:for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			// End of the document, we're done
			break infinitefor

		case tokenType == html.StartTagToken:
			t := tokenizer.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, href := getHref(t)
			if !ok {
				continue
			}

			if url, ok := sanitizeUrl(href, domain); ok {
				urls = append(urls, url)
			}
		}
	}
	return urls, nil
}

// Save the page contents (converted to a byte array) to a file in local storage
// Returns whether the page was saved successfully
func savePage(url url.URL, body []byte) bool{
	// TODO: Take save location as a CMD line flag
	rootDir := "/tmp/scraper"

	dirPath := rootDir + "/" + url.Host + url.Path

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		log.Printf("Cannot create directory %s. \nError: %s", dirPath, err)
		return false
	}

	filePath := dirPath + "/index.html"

	err = ioutil.WriteFile(filePath, body, 0777)
	if err != nil {
		log.Printf("Cannot write to file=%s. \nError: %s", filePath, err)
		return false
	}
	return true
}

// crawl could be called multiple times in parallel to increase productivity.
func crawl(urlSet *concurrentStorage, ch chan url.URL){
	for {
		select {
		case u := <- ch:
			if ok := urlSet.add(u); ok {
				log.Printf("Received url=%s", u.String())
				urls, err := scrape(u)
				if err != nil {
					log.Printf("Could not scrape url=%s.\nError: %s", u.String(), err)
					break
				}

				for _, url := range urls {
					go 	func() {ch <- url}()
				}
			}
		}
	}
}


// TODO: Add 1 sec delay before page makes request

func main() {
	var domain string
	var timeout int

	ch := make(chan url.URL, 2)

	flag.StringVar(&domain, "host", "example.com", "url to scrape")
	flag.IntVar(&timeout, "t", 5, "timeout")
	flag.Parse()

	targetURL, err:= url.Parse(domain)
	if err != nil {
		log.Fatal("Could not parse target url = " + domain)
	}

	if targetURL.Host == "" {
		log.Fatal(" Try the format https://www.example.com. No host found in " + domain)
	}

	// TODO: write function to find a valid schema by requesting with multiple versions of the url
	if targetURL.Scheme == "" {
		targetURL.Scheme = "https"
	}

	urlSet := newConcurrentStorage(targetURL.Host)

	go crawl(urlSet, ch)
	go crawl(urlSet, ch)

	ch <- *targetURL

	time.Sleep(time.Duration(timeout) * time.Second)
	//log.Printf("total in %d seconds = %d", timeout, urlSet.size())
}

/* TODO:
	1. IP simulation
	2. Randomize wait times between each request
	3. User Agent rotation and spoofing
	4. Respect robots.txt
	5. Avoid Honeypots:
		- When following links always take care that the link has proper visibility with no nofollow tag
		- or CSS style display:none
		- color disguised to blend in with the page’s background color
	6. Incorporate some random clicks on the page, mouse movements and random actions that will make a spider looks like a human
*/
