package main

import (
	. "fmt"
	// "io/ioutil"
	"net/http"

	"golang.org/x/net/html"
)

func main() {
	resp, _ := http.Get("https://www.walmart.com/")

	defer resp.Body.Close()

	htmlTokens := html.NewTokenizer(resp.Body)

	for {
		currentTokenType := htmlTokens.Next()

		switch currentTokenType {
		case html.ErrorToken:
			return

		case html.StartTagToken:
			token := htmlTokens.Token()

			if token.Data == "a" {
				for _, a := range token.Attr {
					if a.Key == "href" {
						Println(a.Val)
					}
				}
			}

		}
	}

}
