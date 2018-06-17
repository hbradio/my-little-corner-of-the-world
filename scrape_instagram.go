package main

import (
	"fmt"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type imageUrlsPayload struct {
	Urls []string `json:"urls"`
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var imageUrls []string = scrape("mollyrose30")
	var payload imageUrlsPayload
	payload.Urls = imageUrls
	var jsonBytes []byte
	jsonBytes, _ = json.Marshal(payload)
	fmt.Print(string(jsonBytes))
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:       string(jsonBytes),
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}

// found in https://www.instagram.com/static/bundles/en_US_Commons.js/68e7390c5938.js
// included from profile page
const instagramQueryId = "42323d64886122307be10013ad2dcc45"

// "id": user id, "after": end cursor
const nextPageURL string = `https://www.instagram.com/graphql/query/?query_hash=%s&variables=%s`
const nextPagePayload string = `{"id":"%s","first":50,"after":"%s"}`

var requestID string

type pageInfo struct {
	EndCursor string `json:"end_cursor"`
	NextPage  bool   `json:"has_next_page"`
}

type mainPageData struct {
	Rhxgis    string `json:"rhx_gis"`
	EntryData struct {
		ProfilePage []struct {
			Graphql struct {
				User struct {
					Id    string `json:"id"`
					Media struct {
						Edges []struct {
							Node struct {
								ImageURL     string `json:"display_url"`
								ThumbnailURL string `json:"thumbnail_src"`
								IsVideo      bool   `json:"is_video"`
								Date         int    `json:"date"`
								Dimensions   struct {
									Width  int `json:"width"`
									Height int `json:"height"`
								} `json:"dimensions"`
							} `json::node"`
						} `json:"edges"`
						PageInfo pageInfo `json:"page_info"`
					} `json:"edge_owner_to_timeline_media"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
}

func scrape(instagramAccount string) []string {

	var imageUrls []string

	var actualUserId string
	outputDir := fmt.Sprintf("./instagram_%s/", instagramAccount)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Referrer", "https://www.instagram.com/"+instagramAccount)
		if r.Ctx.Get("gis") != "" {
			gis := fmt.Sprintf("%s:%s", r.Ctx.Get("gis"), r.Ctx.Get("variables"))
			h := md5.New()
			h.Write([]byte(gis))
			gisHash := fmt.Sprintf("%x", h.Sum(nil))
			r.Headers.Set("X-Instagram-GIS", gisHash)
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			idStart := bytes.Index(r.Body, []byte(`:n},queryId:"`))
			requestID = string(r.Body[idStart+13 : idStart+45])
		})
		requestIDURL := e.Request.AbsoluteURL(e.ChildAttr(`link[as="script"]`, "href"))
		d.Visit(requestIDURL)

		dat := e.ChildText("body > script:first-of-type")
		jsonData := dat[strings.Index(dat, "{") : len(dat)-1]
		data := &mainPageData{}
		err := json.Unmarshal([]byte(jsonData), data)
		if err != nil {
			log.Fatal(err)
		}

		os.MkdirAll(outputDir, os.ModePerm)
		page := data.EntryData.ProfilePage[0]
		actualUserId = page.Graphql.User.Id
		for _, obj := range page.Graphql.User.Media.Edges {
			// skip videos
			if obj.Node.IsVideo {
				continue
			}
			fmt.Println("found image:", obj.Node.ThumbnailURL)
			imageUrls = append(imageUrls, obj.Node.ThumbnailURL)
		}

	})

	c.Visit("https://instagram.com/" + instagramAccount)
	return imageUrls
}