package main

import (
	"testing"
	"github.com/aws/aws-lambda-go/events"
)

func TestScraping(t *testing.T) {
	infos := scrape("mollyrose30")
	println(infos)
}

func TestHandler(t *testing.T) {
	var request events.APIGatewayProxyRequest
	handler(request)
}

