package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Printf("Error building request: %s", err)
		return nil, err
	}

	req.Header.Add("User-Agent", "gator")

	c := http.Client{}
	resp, err := (&c).Do(req)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
		return nil, err
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error processing response body: %s", err)
		return nil, err
	}

	rf := RSSFeed{}

	err = xml.Unmarshal(bodyData, &rf)
	if err != nil {
		fmt.Printf("Error unmarshalling xml: %s", err)
		return nil, err
	}

	return &rf, nil
}
