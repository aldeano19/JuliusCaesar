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
	"errors"
)

func Scrape(u url.URL) ([]url.URL, error) {

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

func sanitizeUrl(href string, domain string) (url.URL, bool){
	if strings.Trim(href, " ") == ""{
		return url.URL{}, false
	}

	u, err := url.Parse(href)
	if err != nil {
		// TODO: log event
		return url.URL{}, false
	}

	if u.Host == ""{
		u.Host = domain
	} else if u.Host != domain || u.Path == "/" || u.Path == ""{
		// TODO: Log event
		return url.URL{}, false
	}

	if u.Scheme == ""{
		u.Scheme = "https"
	}

	return *u, true
}

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

func getHttp(url url.URL) (io.ReadCloser, error) {
	resp, err := http.Get(url.String())
	if err != nil {
		log.Printf("HTTP failed to GET url=%s. error=%s\n", url, err)
		return nil, err
	}

	return resp.Body, nil
}

// Get only urls of the specified domain given the body if the page
func getUrls(body []byte, domain string) ([]url.URL, error) {

	// holds only valid urls
	var urls []url.URL

	reader := bytes.NewReader(body)
	tokenizer := html.NewTokenizer(reader)

	for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			// End of the document, we're done
			return urls, nil

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

			url, ok := sanitizeUrl(href, domain)
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func savePage(url url.URL, body []byte){
	rootDir := "/tmp/scraper"

	dirPath := rootDir + url.Path

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		log.Printf("Cannot create directory %s. \nError: %s", dirPath, err)
	}

	filePath := dirPath + "/index.html"
	//outFile, err := os.Create(filePath)
	//if err != nil {
	//	log.Printf("Cannot create file %s. \nError: %s", dirPath, err)
	//	return
	//}
	//defer outFile.Close()

	err = ioutil.WriteFile(filePath, body, 0644)
	if err != nil {
		log.Printf("Cannot write to file=%s. \nError: %s", filePath, err)
		return
	}

	//log.Printf("Saved %s to %s", url.String(), dirPath)
}

func crawl(urlSet concurrentStorage, ch chan url.URL){
	for {
		select {
		case u := <- ch:
			if u, ok := urlSet.Add(u); ok {
				log.Printf("Received url=%s", u.String())
				urls, err := Scrape(u)
				if err != nil {
					log.Printf("Could not scrape url=%s.\nError: %s", u.String(), err)
					break
				}

				for _, url := range urls {
					go func() {ch <- url}()
				}
			}
		}
	}
}