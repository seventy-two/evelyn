package nfl

// scoreboard is the subset of ESPN's NFL scoreboard response that we use.
type scoreboard struct {
	Events []struct {
		Date         string `json:"date"`
		Competitions []struct {
			Competitors []struct {
				HomeAway string `json:"homeAway"`
				Score    string `json:"score"`
				Team     struct {
					ID               string `json:"id"`
					ShortDisplayName string `json:"shortDisplayName"`
				} `json:"team"`
			} `json:"competitors"`
			Status struct {
				Type struct {
					State       string `json:"state"`
					ShortDetail string `json:"shortDetail"`
				} `json:"type"`
			} `json:"status"`
			Situation *struct {
				DownDistanceText string `json:"downDistanceText"`
				Possession       string `json:"possession"`
			} `json:"situation"`
			Broadcasts []struct {
				Names []string `json:"names"`
			} `json:"broadcasts"`
		} `json:"competitions"`
	} `json:"events"`
}
