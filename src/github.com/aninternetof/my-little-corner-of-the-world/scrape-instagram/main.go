package main

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
	"encoding/json"
	"log"
)

type photoInfo struct {
	ThumbnailUrl string `json:"thumbnail_url"`
	Shortcode string `json:"shortcode"`
	Description string `json:"description"`
}


func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	photoInfos := scrape("mollyrose30")
	jsonBytes, err := json.Marshal(photoInfos)
	if err != nil {
		log.Fatal(err)
	}
	println(string(jsonBytes))
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(jsonBytes),
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}

type pageInfo struct {
	EndCursor string `json:"end_cursor"`
	NextPage  bool   `json:"has_next_page"`
}
// var alt = node.edge_media_to_caption.edges[0].node.text
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
								Shortcode    string `json:"shortcode"`
								EdgeMediaToCaption struct {
									Edges []struct {
										Node struct {
											Text string `json:"text"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"edge_media_to_caption"`
								IsVideo      bool   `json:"is_video"`
								Date         int    `json:"date"`
								Dimensions   struct {
									Width  int `json:"width"`
									Height int `json:"height"`
								} `json:"dimensions"`
							} `json:"node"`
						} `json:"edges"`
						PageInfo pageInfo `json:"page_info"`
					} `json:"edge_owner_to_timeline_media"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
}

func scrape(instagramAccount string) []photoInfo {

	var jsonData string
	var photoInfos []photoInfo

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Handling request...")
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
		requestIDURL := e.Request.AbsoluteURL(e.ChildAttr(`link[as="script"]`, "href"))
		d.Visit(requestIDURL)

		dat := e.ChildText("body > script:first-of-type")
		jsonData = dat[strings.Index(dat, "{") : len(dat)-1]

		data := &mainPageData{}
		err := json.Unmarshal([]byte(jsonData), data)
		if err != nil {
			log.Fatal(err)
		}

		page := data.EntryData.ProfilePage[0]
		for _, obj := range page.Graphql.User.Media.Edges {
			// skip videos
			if obj.Node.IsVideo {
				continue
			}
			var info photoInfo
			info.ThumbnailUrl = obj.Node.ThumbnailURL
			info.Shortcode = obj.Node.Shortcode
			info.Description = obj.Node.EdgeMediaToCaption.Edges[0].Node.Text
			photoInfos = append(photoInfos, info)
		}
	})

	c.Visit("https://instagram.com/" + instagramAccount)
	return photoInfos
}
