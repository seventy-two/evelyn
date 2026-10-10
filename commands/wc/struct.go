package wc

import "time"

type WorldCup struct {
	EventGroups []struct {
		DisplayLabel    string `json:"displayLabel"`
		SecondaryGroups []struct {
			DisplayLabel string `json:"displayLabel"`
			Events       []struct {
				Home struct {
					ID            string `json:"id"`
					FullName      string `json:"fullName"`
					ShortName     string `json:"shortName"`
					Urn           string `json:"urn"`
					RunningScores struct {
						Halftime string `json:"halftime"`
						Fulltime string `json:"fulltime"`
					} `json:"runningScores"`
					ScoreUnconfirmed string `json:"scoreUnconfirmed"`
					Actions          []struct {
						PlayerUrn  string `json:"playerUrn"`
						PlayerName string `json:"playerName"`
						ActionType string `json:"actionType"`
						Actions    []struct {
							Type      string `json:"type"`
							TypeLabel struct {
								Value      string `json:"value"`
								Accessible string `json:"accessible"`
							} `json:"typeLabel"`
							TimeLabel struct {
								Value      string `json:"value"`
								Accessible string `json:"accessible"`
							} `json:"timeLabel"`
						} `json:"actions"`
					} `json:"actions"`
					Score string `json:"score"`
				} `json:"home"`
				Away struct {
					ID            string `json:"id"`
					FullName      string `json:"fullName"`
					ShortName     string `json:"shortName"`
					Urn           string `json:"urn"`
					RunningScores struct {
						Halftime string `json:"halftime"`
						Fulltime string `json:"fulltime"`
					} `json:"runningScores"`
					ScoreUnconfirmed string `json:"scoreUnconfirmed"`
					Actions          []struct {
						PlayerUrn  string `json:"playerUrn"`
						PlayerName string `json:"playerName"`
						ActionType string `json:"actionType"`
						Actions    []struct {
							Type      string `json:"type"`
							TypeLabel struct {
								Value      string `json:"value"`
								Accessible string `json:"accessible"`
							} `json:"typeLabel"`
							TimeLabel struct {
								Value      string `json:"value"`
								Accessible string `json:"accessible"`
							} `json:"timeLabel"`
						} `json:"actions"`
					} `json:"actions"`
					Score string `json:"score"`
				} `json:"away"`
				ID                 string    `json:"id"`
				Urn                string    `json:"urn"`
				EventGroupingLabel string    `json:"eventGroupingLabel"`
				StartDateTime      time.Time `json:"startDateTime"`
				TournamentID       string    `json:"tournamentId"`
				Date               struct {
					Iso        time.Time `json:"iso"`
					Time       string    `json:"time"`
					ShortDate  string    `json:"shortDate"`
					LongDate   string    `json:"longDate"`
					DayOfWeek  string    `json:"dayOfWeek"`
					Day        string    `json:"day"`
					Month      string    `json:"month"`
					ShortMonth string    `json:"shortMonth"`
					Year       string    `json:"year"`
					IsoDate    string    `json:"isoDate"`
				} `json:"date"`
				PeriodLabel struct {
					Value      string `json:"value"`
					Accessible string `json:"accessible"`
				} `json:"periodLabel"`
				Time struct {
					AccessibleTime string `json:"accessibleTime"`
					DisplayTimeUK  string `json:"displayTimeUK"`
					TimeCertainty  bool   `json:"timeCertainty"`
				} `json:"time"`
				Status        string `json:"status"`
				StatusComment struct {
					Value      string `json:"value"`
					Accessible string `json:"accessible"`
				} `json:"statusComment"`
				Participants []struct {
					ID   string `json:"id"`
					Urn  string `json:"urn"`
					Name struct {
						FullName  string `json:"fullName"`
						ShortName string `json:"shortName"`
					} `json:"name"`
					Score         string `json:"score"`
					HalftimeScore string `json:"halftimeScore"`
					FulltimeScore string `json:"fulltimeScore"`
					Alignment     string `json:"alignment"`
				} `json:"participants"`
				HeadToHeadDetailLabel string `json:"headToHeadDetailLabel"`
				Tournament            struct {
					ID                string `json:"id"`
					Name              string `json:"name"`
					DisambiguatedName string `json:"disambiguatedName"`
					Urn               string `json:"urn"`
					ThingsGUID        string `json:"thingsGuid"`
				} `json:"tournament"`
				Stage struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Urn  string `json:"urn"`
				} `json:"stage"`
				Round struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Urn  string `json:"urn"`
				} `json:"round"`
				Winner                 string `json:"winner"`
				TipoTopicID            string `json:"tipoTopicId"`
				OnwardJourneyLink      string `json:"onwardJourneyLink"`
				AccessibleEventSummary string `json:"accessibleEventSummary"`
				HasStandings           bool   `json:"hasStandings"`
			} `json:"events"`
		} `json:"secondaryGroups"`
	} `json:"eventGroups"`
	Sport                   string `json:"sport"`
	SelectedStartDate       string `json:"selectedStartDate"`
	SelectedEndDate         string `json:"selectedEndDate"`
	ShouldShowScorersButton bool   `json:"shouldShowScorersButton"`
	Urn                     string `json:"urn"`
	MaximumScoreDigits      int    `json:"maximumScoreDigits"`
}
