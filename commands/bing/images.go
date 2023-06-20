package bing

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (c *Client) GetThumbnail(input string) string {
	resp := c.bing(input)
	if len(resp.Value) > 0 {
		return resp.Value[0].ThumbnailURL
	}
	return ""
}

func (c *Client) GetImage(input []string) string {
	resp := c.bing(strings.Join(input, "+"))
	if len(resp.Value) > 0 {
		return resp.Value[0].ContentURL
	}
	return ""
}

func (c *Client) bing(input string) *response {
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", serviceConfig.TargetURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Set the query parameters
	q := req.URL.Query()
	q.Set("q", input)
	q.Set("count", "1") // Retrieve only 1 image
	req.URL.RawQuery = q.Encode()

	// Set the subscription key in the request headers
	req.Header.Set("Ocp-Apim-Subscription-Key", serviceConfig.APIKey)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %d", resp.StatusCode)
	}

	r := &response{}

	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

// Parse the response JSON
type response struct {
	Value []struct {
		ThumbnailURL string `json:"thumbnailUrl"`
		ContentURL   string `json:"contentUrl"`
	} `json:"value"`
}
