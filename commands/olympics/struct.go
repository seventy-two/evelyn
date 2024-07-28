package olympics

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
