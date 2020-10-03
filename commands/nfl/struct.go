package nfl

type nflResponse struct {
	Game game `json:"-"`
}

type game struct {
	Home struct {
		Score struct {
			Num1 int `json:"1"`
			Num2 int `json:"2"`
			Num3 int `json:"3"`
			Num4 int `json:"4"`
			Num5 int `json:"5"`
			T    int `json:"T"`
		} `json:"score"`
		Abbr string `json:"abbr"`
		To   int    `json:"to"`
	} `json:"home"`
	Away struct {
		Score struct {
			Num1 int `json:"1"`
			Num2 int `json:"2"`
			Num3 int `json:"3"`
			Num4 int `json:"4"`
			Num5 int `json:"5"`
			T    int `json:"T"`
		} `json:"score"`
		Abbr string `json:"abbr"`
		To   int    `json:"to"`
	} `json:"away"`
	Bp      int         `json:"bp"`
	Down    int         `json:"down"`
	Togo    int         `json:"togo"`
	Clock   string      `json:"clock"`
	Posteam string      `json:"posteam"`
	Note    interface{} `json:"note"`
	Redzone bool        `json:"redzone"`
	Stadium string      `json:"stadium"`
	Media   struct {
		Radio struct {
			Home interface{} `json:"home"`
			Away interface{} `json:"away"`
		} `json:"radio"`
		Tv    string      `json:"tv"`
		Sat   interface{} `json:"sat"`
		Sathd interface{} `json:"sathd"`
	} `json:"media"`
	Yl  string  `json:"yl"`
	Qtr *string `json:"qtr"`
}
