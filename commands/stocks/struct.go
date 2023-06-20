package stocks

import "time"

type Quote struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}

type Lookup struct {
	BestMatches []struct {
		Symbol      string `json:"1. symbol"`
		Name        string `json:"2. name"`
		Type        string `json:"3. type"`
		Region      string `json:"4. region"`
		MarketOpen  string `json:"5. marketOpen"`
		MarketClose string `json:"6. marketClose"`
		Timezone    string `json:"7. timezone"`
		Currency    string `json:"8. currency"`
		MatchScore  string `json:"9. matchScore"`
	} `json:"bestMatches"`
}

type Earnings struct {
	Symbol           string
	Name             string
	ReportDate       time.Time
	FiscalDateEnding time.Time
	Estimate         float32
	Currency         string
}

type ExchangeRate struct {
	RealtimeCurrencyExchangeRate struct {
		FromCurrencyCode string `json:"1. From_Currency Code"`
		FromCurrencyName string `json:"2. From_Currency Name"`
		ToCurrencyCode   string `json:"3. To_Currency Code"`
		ToCurrencyName   string `json:"4. To_Currency Name"`
		ExchangeRate     string `json:"5. Exchange Rate"`
		LastRefreshed    string `json:"6. Last Refreshed"`
		TimeZone         string `json:"7. Time Zone"`
		BidPrice         string `json:"8. Bid Price"`
		AskPrice         string `json:"9. Ask Price"`
	} `json:"Realtime Currency Exchange Rate"`
}

type CNBCStock struct {
	FormattedQuoteResult struct {
		FormattedQuote []struct {
			Symbol             string `json:"symbol"`
			SymbolType         string `json:"symbolType"`
			Code               int    `json:"code"`
			Name               string `json:"name"`
			ShortName          string `json:"shortName"`
			OnAirName          string `json:"onAirName"`
			AltName            string `json:"altName"`
			Last               string `json:"last"`
			LastTimedate       string `json:"last_timedate"`
			LastTime           string `json:"last_time"`
			Changetype         string `json:"changetype"`
			Type               string `json:"type"`
			SubType            string `json:"subType"`
			Exchange           string `json:"exchange"`
			Source             string `json:"source"`
			Open               string `json:"open"`
			High               string `json:"high"`
			Low                string `json:"low"`
			Change             string `json:"change"`
			ChangePct          string `json:"change_pct"`
			CurrencyCode       string `json:"currencyCode"`
			Volume             string `json:"volume"`
			VolumeAlt          string `json:"volume_alt"`
			Provider           string `json:"provider"`
			PreviousDayClosing string `json:"previous_day_closing"`
			AltSymbol          string `json:"altSymbol"`
			RealTime           string `json:"realTime"`
			Curmktstatus       string `json:"curmktstatus"`
			Pe                 string `json:"pe"`
			MktcapView         string `json:"mktcapView"`
			Dividend           string `json:"dividend"`
			Dividendyield      string `json:"dividendyield"`
			Beta               string `json:"beta"`
			Tendayavgvol       string `json:"tendayavgvol"`
			Pcttendayvol       string `json:"pcttendayvol"`
			Yrhiprice          string `json:"yrhiprice"`
			Yrhidate           string `json:"yrhidate"`
			Yrloprice          string `json:"yrloprice"`
			Yrlodate           string `json:"yrlodate"`
			Eps                string `json:"eps"`
			Sharesout          string `json:"sharesout"`
			Revenuettm         string `json:"revenuettm"`
			Fpe                string `json:"fpe"`
			Feps               string `json:"feps"`
			Psales             string `json:"psales"`
			Fsales             string `json:"fsales"`
			Fpsales            string `json:"fpsales"`
			Streamable         string `json:"streamable"`
			IssueID            string `json:"issue_id"`
			IssuerID           string `json:"issuer_id"`
			CountryCode        string `json:"countryCode"`
			TimeZone           string `json:"timeZone"`
			FeedSymbol         string `json:"feedSymbol"`
			Portfolioindicator string `json:"portfolioindicator"`
			Roettm             string `json:"ROETTM"`
			Netprofttm         string `json:"NETPROFTTM"`
			Grosmgnttm         string `json:"GROSMGNTTM"`
			Ttmebitd           string `json:"TTMEBITD"`
			Debteqtyq          string `json:"DEBTEQTYQ"`
			ExtendedMktQuote   struct {
				Type         string `json:"type"`
				Source       string `json:"source"`
				Last         string `json:"last"`
				LastTimedate string `json:"last_timedate"`
				LastTime     string `json:"last_time"`
				Change       string `json:"change"`
				ChangePct    string `json:"change_pct"`
				Volume       string `json:"volume"`
				VolumeAlt    string `json:"volume_alt"`
				Changetype   string `json:"changetype"`
			} `json:"ExtendedMktQuote"`
			EventData struct {
				NextEarningsDate      string `json:"next_earnings_date"`
				NextEarningsDateToday string `json:"next_earnings_date_today"`
				AnnounceTime          string `json:"announce_time"`
				DivExDate             string `json:"div_ex_date"`
				DivExDateToday        string `json:"div_ex_date_today"`
				DivAmount             string `json:"div_amount"`
				Yrhiind               string `json:"yrhiind"`
				Yrloind               string `json:"yrloind"`
				IsHalted              string `json:"is_halted"`
			} `json:"EventData"`
		} `json:"FormattedQuote"`
	} `json:"FormattedQuoteResult"`
}
