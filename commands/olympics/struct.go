package olympics

import "time"

type Olympics struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Year               int    `json:"year"`
	CompetitionSummary struct {
		EventsTotal         int `json:"eventsTotal"`
		EventsScheduled     int `json:"eventsScheduled"`
		EventsComplete      int `json:"eventsComplete"`
		AwardedGoldMedals   int `json:"awardedGoldMedals"`
		AwardedSilverMedals int `json:"awardedSilverMedals"`
		AwardedBronzeMedals int `json:"awardedBronzeMedals"`
	} `json:"competitionSummary"`
	Disciplines []any `json:"disciplines"`
	Medals      []struct {
		ID      string `json:"id"`
		Country string `json:"country"`
		Gold    int    `json:"gold"`
		Silver  int    `json:"silver"`
		Bronze  int    `json:"bronze"`
		Total   int    `json:"total"`
	} `json:"medals"`
	LastUpdated int64 `json:"lastUpdated"`
}

type Events struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Year   int    `json:"year"`
	Events []struct {
		Discipline string    `json:"discipline"`
		Name       string    `json:"name"`
		Gender     string    `json:"gender"`
		StartTime  time.Time `json:"startTime"`
		EndTime    time.Time `json:"endTime"`
		Gold       []struct {
			ID             string `json:"id"`
			Country        string `json:"country"`
			CompetitorName string `json:"competitorName"`
		} `json:"gold"`
		Silver []struct {
			ID             string `json:"id"`
			Country        string `json:"country"`
			CompetitorName string `json:"competitorName"`
		} `json:"silver"`
		Bronze []struct {
			ID             string `json:"id"`
			Country        string `json:"country"`
			CompetitorName string `json:"competitorName"`
		} `json:"bronze"`
	} `json:"events"`
}
