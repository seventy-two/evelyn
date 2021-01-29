package stocks

type IEXStocks struct {
	Symbol                string  `json:"symbol"`
	CompanyName           string  `json:"companyName"`
	CalculationPrice      string  `json:"calculationPrice"`
	Open                  float64 `json:"open"`
	OpenTime              int64   `json:"openTime"`
	Close                 float64 `json:"close"`
	CloseTime             int64   `json:"closeTime"`
	High                  float64 `json:"high"`
	Low                   float64 `json:"low"`
	LatestPrice           float64 `json:"latestPrice"`
	LatestSource          string  `json:"latestSource"`
	LatestTime            string  `json:"latestTime"`
	LatestUpdate          int64   `json:"latestUpdate"`
	LatestVolume          int     `json:"latestVolume"`
	IexRealtimePrice      float64 `json:"iexRealtimePrice"`
	IexRealtimeSize       int     `json:"iexRealtimeSize"`
	IexLastUpdated        int64   `json:"iexLastUpdated"`
	DelayedPrice          float64 `json:"delayedPrice"`
	DelayedPriceTime      int64   `json:"delayedPriceTime"`
	ExtendedPrice         float64 `json:"extendedPrice"`
	ExtendedChange        float64 `json:"extendedChange"`
	ExtendedChangePercent float64 `json:"extendedChangePercent"`
	ExtendedPriceTime     int64   `json:"extendedPriceTime"`
	PreviousClose         float64 `json:"previousClose"`
	Change                float64 `json:"change"`
	ChangePercent         float64 `json:"changePercent"`
	IexMarketPercent      float64 `json:"iexMarketPercent"`
	IexVolume             int     `json:"iexVolume"`
	AvgTotalVolume        int     `json:"avgTotalVolume"`
	IexBidPrice           float64 `json:"iexBidPrice"`
	IexBidSize            int     `json:"iexBidSize"`
	IexAskPrice           float64 `json:"iexAskPrice"`
	IexAskSize            int     `json:"iexAskSize"`
	MarketCap             int64   `json:"marketCap"`
	PeRatio               float64 `json:"peRatio"`
	Week52High            float64 `json:"week52High"`
	Week52Low             float64 `json:"week52Low"`
	YtdChange             float64 `json:"ytdChange"`
}

type Lookup struct {
	ResultSet struct {
		Query  string `json:"Query"`
		Result []struct {
			Symbol   string `json:"symbol"`
			Name     string `json:"name"`
			Exch     string `json:"exch"`
			Type     string `json:"type"`
			ExchDisp string `json:"exchDisp"`
			TypeDisp string `json:"typeDisp"`
		} `json:"Result"`
	} `json:"ResultSet"`
}

type Crypto struct {
	Symbol           string `json:"symbol"`
	PrimaryExchange  string `json:"primaryExchange"`
	Sector           string `json:"sector"`
	CalculationPrice string `json:"calculationPrice"`
	LatestPrice      string `json:"latestPrice"`
	LatestSource     string `json:"latestSource"`
	LatestUpdate     int64  `json:"latestUpdate"`
	LatestVolume     string `json:"latestVolume"`
	BidPrice         string `json:"bidPrice"`
	BidSize          string `json:"bidSize"`
	AskPrice         string `json:"askPrice"`
	AskSize          string `json:"askSize"`
	High             string `json:"high"`
	Low              string `json:"low"`
	PreviousClose    string `json:"previousClose"`
}

type Currency struct {
	Results []struct {
		Symbol    string  `json:"symbol"`
		Rate      float64 `json:"rate"`
		Timestamp int64   `json:"timestamp"`
		IsDerived bool    `json:"isDerived"`
	}
}

type CurrencyRates struct {
	Rates struct {
		CAD float64 `json:"CAD"`
		HKD float64 `json:"HKD"`
		ISK float64 `json:"ISK"`
		PHP float64 `json:"PHP"`
		DKK float64 `json:"DKK"`
		HUF float64 `json:"HUF"`
		CZK float64 `json:"CZK"`
		GBP float64 `json:"GBP"`
		RON float64 `json:"RON"`
		SEK float64 `json:"SEK"`
		IDR float64 `json:"IDR"`
		INR float64 `json:"INR"`
		BRL float64 `json:"BRL"`
		RUB float64 `json:"RUB"`
		HRK float64 `json:"HRK"`
		JPY float64 `json:"JPY"`
		THB float64 `json:"THB"`
		CHF float64 `json:"CHF"`
		EUR float64 `json:"EUR"`
		MYR float64 `json:"MYR"`
		BGN float64 `json:"BGN"`
		TRY float64 `json:"TRY"`
		CNY float64 `json:"CNY"`
		NOK float64 `json:"NOK"`
		NZD float64 `json:"NZD"`
		ZAR float64 `json:"ZAR"`
		USD float64 `json:"USD"`
		MXN float64 `json:"MXN"`
		SGD float64 `json:"SGD"`
		AUD float64 `json:"AUD"`
		ILS float64 `json:"ILS"`
		KRW float64 `json:"KRW"`
		PLN float64 `json:"PLN"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}
