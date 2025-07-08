package generation

import (
	"testing"
	"time"
)

func TestGetDuration(t *testing.T) {
	youtube := &Youtube{
		Items: []struct {
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
		}{
			{
				ContentDetails: struct {
					Duration        string `json:"duration"`
					Dimension       string `json:"dimension"`
					Definition      string `json:"definition"`
					Caption         string `json:"caption"`
					LicensedContent bool   `json:"licensedContent"`
					ContentRating   struct {
					} `json:"contentRating"`
					Projection string `json:"projection"`
				}{
					Duration: "PT14M34S",
				},
			},
		},
	}

	expectedDuration := 14*time.Minute + 34*time.Second
	duration := youtube.GetDuration()

	if duration != expectedDuration {
		t.Errorf("expected %v, got %v", expectedDuration, duration)
	}
}
