package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gocolly/colly"

	_ "gopkg.in/cq.v1"
)

type company struct {
	Name    string    `json:"name"`
	Symbol  string    `json:"symbol"`
	URL     string    `json:"url"`
	Price   string    `json:"price"`
	Related []company `json:"related"`
}

func (c company) String() string {
	return fmt.Sprintf("name: %s, url: %s, price: %s \n", c.Name, c.URL, c.Price)
}

func main() {

	db, err := sql.Open("neo4j-cypher", "http://localhost:7474")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connection established with db")
	}
	defer db.Close()
	firstChannel := make(chan bool)
	scrape("Apple", "http://markets.businessinsider.com/stocks/aapl-stock", firstChannel)
	<-firstChannel

}

func scrape(symbolOrigin string, href string, pChannel chan bool) {
	var session []company
	c := colly.NewCollector()

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		namecomp := e.ChildText("a[href][title]")
		newcompany := company{Name: namecomp}
		newcompany.URL = e.ChildAttr("a", "href")
		newcompany.Price = e.ChildText("td[width='15%']")
		if newcompany.URL != "" && newcompany.Name != "" && newcompany.Price != "" {
			session = append(session, newcompany)
		}

		statement := "{\n  \"statements\" : [ {\n    \"statement\" : \"MERGE (n:Company {name:'" + symbolOrigin + "'}) MERGE (test2:Company {name:'" + newcompany.Name + "'}) MERGE (n)-[:isin]->(test2)\"\n\t}\n]\n}"
		mychannel := make(chan bool)
		go postData(statement, mychannel)
		<-mychannel
		mychannel2 := make(chan bool)
		go scrape(symbolOrigin, "http://markets.businessinsider.com"+newcompany.URL, mychannel2)
		<-mychannel2
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(href)
}

func postData(request string, mychannel chan bool) {
	url := "http://localhost:7474/db/data/transaction/commit"
	payload := strings.NewReader(request)
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Basic bmVvNGo6aWduYWNpbw==")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("postman-token", "a15b1e94-b8ab-d8ae-3046-37702396bae3")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	mychannel <- true
}
