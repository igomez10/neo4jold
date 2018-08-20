package main

import (
	"encoding/json"
)

type WikiApiResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Warnings      struct {
		Links struct {
			NAMING_FAILED string `json:"*"`
		} `json:"links"`
	} `json:"warnings"`
	Query struct {
		Normalized []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"normalized"`
		Pages json.RawMessage
	} `json:"query"`
}

type data struct {
	Pageid int    `json:"pageid"`
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
	Links  []struct {
		Ns    int    `json:"ns"`
		Title string `json:"title"`
	} `json:"links"`
}

type info map[string]data
