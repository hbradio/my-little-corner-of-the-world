package main

import "testing"

func TestScraping(t *testing.T) {
	var json string = scrape("mollyrose30")
	println(json)
}

