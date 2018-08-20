package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "gopkg.in/cq.v1"
)

func main() {

	donePages := make(chan bool)
	chanPages := make(chan string, 1000)

	doneStatements := make(chan bool)
	chanStatements := make(chan string, 1000)

	db, _ := sql.Open("neo4j-cypher", "http://neo4j:nacho@localhost:7474")
	defer db.Close()

	go func() {
		for j := range chanStatements {
			go func(j string) {
				result, err := db.Exec(j)
				if err != nil {
					fmt.Printf("Error posting to neo4j %s \n", err)
				} else {
					fmt.Println("Success:", result)
				}
			}(j)
		}
		doneStatements <- true
	}()

	go func() {
		for j := range chanPages {
			go getWikiInfo(j, chanPages, chanStatements, false)
		}
		donePages <- true
	}()

	page := "Anexo:Congresistas_colombianos_2014-2018"
	getWikiInfo(page, chanPages, chanStatements, true)

	<-donePages
	close(chanPages)

}
func getWikiInfo(pageTitle string, c1 chan string, c2 chan string, initial bool) {
	fmt.Printf("Exploring %s \n", pageTitle)
	baseURL := "http://es.wikipedia.org/w/api.php?action=query&format=json&prop=links&titles=%s&pllimit=5000&plnamespace="
	url := fmt.Sprintf(baseURL, pageTitle)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
	}

	var war WikiApiResponse
	err = json.Unmarshal(body, &war)
	if err != nil {
		log.Println("Unable to parse json", err)
	}

	var dataMap info
	err = json.Unmarshal(war.Query.Pages, &dataMap)

	var dataI data
	for _, v := range dataMap {
		dataI = v
	}

	for _, v := range dataI.Links {
		title := strings.Replace(v.Title, " ", "_", -1)
		// fmt.Printf("%d. %s \n", i, title)
		if initial {
			c1 <- title
			statement := fmt.Sprintf(`
		MERGE (a:Page {title: '%s', name: '%s'})`, strings.Replace(pageTitle, "'", `\'`, -1), strings.Replace(pageTitle, "_", " ", -1))
			c2 <- statement
		} else {
			statement := fmt.Sprintf(`
		MERGE (a:Page {title: '%s', name: '%s'})
		MERGE (b:Page {title: '%s', name: '%s'})
		MERGE (a)-[:SHOWS_IN]->(b)`,
				strings.Replace(pageTitle, "'", `\'`, -1),
				strings.Replace(pageTitle, "_", " ", -1),
				strings.Replace(title, "'", `\'`, -1),
				strings.Replace(v.Title, "'", `\'`, -1))
			c2 <- statement
		}
	}
}

func init() {
	db, err := sql.Open("neo4j-cypher", "http://neo4j:nacho@localhost:7474")
	if err != nil {
		fmt.Println("Error establishing connection with neo4j")
		fmt.Println(err)
		os.Exit(1)
	}
	db.Close()
}
