package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	url := "https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html"
	resp, _ := http.Get(url)
	doc := html.NewTokenizer(resp.Body)

	for tokenType := doc.Next(); tokenType != html.ErrorToken; { //create token type, the for executes if tokenType != htmlErrorToken
		// current token
		token := doc.Token()

		if tokenType == html.StartTagToken { //entering new tag
			if token.DataAtom != atom.A { // token.DataAtom contains the name of the tag (script, section, div, h1 etc)
				tokenType = doc.Next() // token type is either Text, Endtag, StartTag
				continue
			}
			for index, attr := range token.Attr { // attr is every possible attr in the html tag. it can be href title target etc
				if attr.Key == "href" { // we are looking for attr that match with href
					fmt.Println("LINK FOUND", attr)
				}
			}
		}
		tokenType = doc.Next()
	}

	resp.Body.Close()
}
