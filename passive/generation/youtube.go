package generation

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/seventy-two/Cara/web"
)

func getYouTube(query string) (*Youtube, error) {
	results := &Youtube{}
	u, err := url.Parse(query)
	if err != nil {
		return nil, err
	}
	queryParams := u.Query()
	videoID := queryParams.Get("v")
	if videoID == "" {
		// Check if the URL is in the short format
		if strings.Contains(u.Host, "youtu.be") {
			videoID = strings.TrimPrefix(u.Path, "/")
			// Remove any query parameters from the video ID
			if idx := strings.Index(videoID, "?"); idx != -1 {
				videoID = videoID[:idx]
			}
		}
		if videoID == "" {
			return nil, fmt.Errorf("no video ID found in URL")
		}
	}
	q := fmt.Sprintf(serviceConfig.YoutubeURL, videoID, serviceConfig.YoutubeAPIKey)
	err = web.GetJSON(q, results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

type Youtube struct {
	Kind  string `json:"kind"`
	Etag  string `json:"etag"`
	Items []struct {
		Kind           string `json:"kind"`
		Etag           string `json:"etag"`
		ID             string `json:"id"`
		ContentDetails struct {
			Duration        string `json:"duration"`
			Dimension       string `json:"dimension"`
			Definition      string `json:"definition"`
			Caption         string `json:"caption"`
			LicensedContent bool   `json:"licensedContent"`
			ContentRating   struct {
			} `json:"contentRating"`
			Projection string `json:"projection"`
		} `json:"contentDetails"`
	} `json:"items"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
}

func (y *Youtube) GetDuration() time.Duration {
	d, err := time.ParseDuration(strings.ToLower(strings.TrimPrefix(y.Items[0].ContentDetails.Duration, "PT")))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return d
}
