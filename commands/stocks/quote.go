package stocks

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func lookupAndGetStock(query string) (*stock, error) {
	symbol, err := lookupStock(query)
	if err != nil {
		fmt.Println(err)
	}
	if symbol == "" {
		symbol = query
	}
	s, err := getStockData(symbol)
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return nil, nil
	}
	return s[0], nil
}

func getStockData(symbol string) ([]*stock, error) {
	url := fmt.Sprintf(serviceConfig.QuoteURL, symbol)
	stocks := &CNBCStock{}
	err := web.GetJSON(url, stocks)
	if err != nil {
		return nil, err
	}
	var stockz []*stock
	for _, s := range stocks.FormattedQuoteResult.FormattedQuote {
		unit := priceUnit(s.CurrencyCode)
		price := withUnit(s.Last, unit)
		change := s.Change
		percentChange := s.ChangePct
		source := s.Source
		time := s.LastTimedate
		extPrice := ""
		extChange := ""
		extPct := ""

		if s.ExtendedMktQuote.LastTime > s.LastTime {
			extPrice = "\n🌙 " + withUnit(s.ExtendedMktQuote.Last, unit)
			extChange = "\n🌙 " + s.ExtendedMktQuote.Change
			extPct = "(" + s.ExtendedMktQuote.ChangePct + ")"
			source = s.ExtendedMktQuote.Source
			time = s.ExtendedMktQuote.LastTimedate
		}
		if s.Name == "" {
			continue
		}
		stockz = append(stockz, &stock{
			symbol:           s.Symbol,
			name:             s.Name,
			price:            price,
			high:             s.High,
			low:              s.Low,
			change:           change,
			percentChange:    percentChange,
			extPrice:         extPrice,
			extChange:        extChange,
			extPercentChange: extPct,
			source:           source,
			eps:              s.Eps,
			pe:               s.Pe,
			mktcap:           s.MktcapView,
			time:             time,
			yearHigh:         s.Yrhiprice,
			yearLow:          s.Yrloprice,
			currency:         s.CurrencyCode,
		})
	}
	return stockz, nil
}

// priceUnit is the suffix shown after prices: "p" for London's pence quotes,
// nothing otherwise.
func priceUnit(currency string) string {
	switch currency {
	case "GBp", "GBX", "GBx":
		return "p"
	}
	return ""
}

// withUnit appends unit to a non-empty price.
func withUnit(price, unit string) string {
	if price == "" {
		return price
	}
	return price + unit
}

// lookupStock turns a ticker or company name into the symbol to quote, using
// CNBC's symbol search so the result is always one the CNBC quote API knows.
// The search matches the start of names literally, so a multi-word query that
// finds nothing is retried hyphenated ("rolls royce" → Rolls-Royce), then as
// its first word.
func lookupStock(query string) (string, error) {
	var lastErr error
	for _, q := range searchVariants(query) {
		var matches []LookupMatch
		if err := web.GetJSON(fmt.Sprintf(serviceConfig.LookupURL, q), &matches); err != nil {
			lastErr = err
			continue
		}
		sym, london := pickSymbol(query, matches)
		if sym == "" {
			continue
		}
		if len(london) > 0 {
			quotes, err := getStockData(strings.Join(london, "|"))
			if err != nil {
				return sym, nil
			}
			if uk := ukListing(quotes); uk != "" {
				return uk, nil
			}
		}
		return sym, nil
	}
	return "", lastErr
}

// ukListing returns the first quote priced in pence or pounds with a price,
// which marks a UK company's home listing. Foreign companies' London lines
// (Apple's 0R2V-GB) trade in their own currency, often with no price.
func ukListing(quotes []*stock) string {
	for _, q := range quotes {
		switch strings.ToUpper(q.currency) {
		case "GBP", "GBX":
			if q.price != "" {
				return q.symbol
			}
		}
	}
	return ""
}

// searchVariants lists the search strings lookupStock tries, in order, for a
// "+"-joined query.
func searchVariants(query string) []string {
	words := strings.FieldsFunc(query, func(r rune) bool { return r == '+' || r == ' ' })
	if len(words) < 2 {
		return []string{query}
	}
	return []string{strings.Join(words, "+"), strings.Join(words, "-"), words[0]}
}

// pickSymbol chooses what to quote from search results. It returns the
// symbol to use and, when the chosen company also trades in London, those
// London symbols, which lookupStock quotes to see if it's a UK company
// (priced in pence) whose home listing should win.
//
// CNBC's ranking can put another company first (Gilead for "astrazeneca",
// NOVONIX for "novo"), so a company whose name starts with the query's words
// is preferred.
//
// An exact ticker wins (so "ANT" stays ANT), though a UK company's US ticker
// ("bp") still gets its London listings checked. Otherwise the company is
// the first stock whose issuer has a US or London listing (so Apple Inc. beats
// a Shanghai "Apple Flavor"), and its symbol is its US stock, else any US
// listing, else its best-ranked one. With no stocks at all, the best US
// listing (an ETF, say), else the best match.
func pickSymbol(query string, matches []LookupMatch) (string, []string) {
	var found []LookupMatch
	for _, m := range matches {
		if m.Symbol != "" {
			found = append(found, m)
		}
	}
	if len(found) == 0 {
		return "", nil
	}
	// Group listings by company; a match without an issuer is its own company.
	company := func(m LookupMatch) string {
		if m.IssuerID != "" {
			return m.IssuerID
		}
		return "symbol:" + m.Symbol
	}
	listings := map[string][]LookupMatch{}
	for _, m := range found {
		listings[company(m)] = append(listings[company(m)], m)
	}
	usOrUK := func(ls []LookupMatch) bool {
		for _, l := range ls {
			if l.CountryCode == "US" || l.CountryCode == "GB" {
				return true
			}
		}
		return false
	}

	londonOf := func(id string) []string {
		var london []string
		for _, l := range listings[id] {
			if l.CountryCode == "GB" && l.IssueType == "STOCK" {
				london = append(london, l.Symbol)
			}
		}
		return london
	}

	q := strings.ReplaceAll(query, "+", " ")
	for _, m := range found {
		if strings.EqualFold(m.Symbol, q) {
			if m.CountryCode == "GB" || m.IssueType != "STOCK" {
				return m.Symbol, nil
			}
			return m.Symbol, londonOf(company(m))
		}
	}

	target := ""
	for _, m := range found {
		if m.IssueType == "STOCK" && usOrUK(listings[company(m)]) && nameStartsWith(m.Name, q) {
			target = company(m)
			break
		}
	}
	if target == "" {
		for _, m := range found {
			if m.IssueType == "STOCK" && usOrUK(listings[company(m)]) {
				target = company(m)
				break
			}
		}
	}
	if target == "" {
		for _, m := range found {
			if m.IssueType == "STOCK" {
				target = company(m)
				break
			}
		}
	}
	if target == "" {
		for _, m := range found {
			if m.CountryCode == "US" {
				return m.Symbol, nil
			}
		}
		return found[0].Symbol, nil
	}

	ls := listings[target]
	london := londonOf(target)
	for _, l := range ls {
		if l.CountryCode == "US" && l.IssueType == "STOCK" {
			return l.Symbol, london
		}
	}
	for _, l := range ls {
		if l.CountryCode == "US" {
			return l.Symbol, london
		}
	}
	return ls[0].Symbol, london
}

// nameStartsWith reports whether name begins with query as whole words,
// ignoring case and punctuation: "Rolls-Royce Holdings" starts with
// "rolls royce", "Novo Nordisk" with "novo", but "NOVONIX" doesn't.
func nameStartsWith(name, query string) bool {
	words := func(s string) string {
		return strings.Join(strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}), " ")
	}
	n, q := words(name), words(query)
	return q != "" && (n == q || strings.HasPrefix(n, q+" "))
}

func outputStock(q *stock, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	img := fmt.Sprintf(serviceConfig.LogoURL, q.symbol, serviceConfig.LogoAPIKey)

	e := &discordgo.MessageEmbed{
		Title:       q.symbol + " - " + q.name,
		Description: q.source + " (" + q.time + ")",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: img,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Price",
				Value:  fmt.Sprintf("%s%s", q.price, q.extPrice),
				Inline: true,
			},
			{
				Name:   "Change",
				Value:  fmt.Sprintf("%s(%s)%s%s", q.change, q.percentChange, q.extChange, q.extPercentChange),
				Inline: true,
			},
		},
	}

	s.ChannelMessageSendEmbed(channelID, e)
}

func outputBigStock(q *stock, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	img := fmt.Sprintf(serviceConfig.LogoURL, q.symbol, serviceConfig.LogoAPIKey)

	e := &discordgo.MessageEmbed{
		Title:       q.symbol + " - " + q.name,
		Description: q.source + " (" + q.time + ")",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: img,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Price",
				Value:  fmt.Sprintf("%s%s", q.price, q.extPrice),
				Inline: true,
			},
			{
				Name:   "Change",
				Value:  fmt.Sprintf("%s(%s)%s%s", q.change, q.percentChange, q.extChange, q.extPercentChange),
				Inline: true,
			},
			{
				Name:   "Market Cap",
				Value:  q.mktcap,
				Inline: true,
			},
			{
				Name:   "High",
				Value:  q.high,
				Inline: true,
			},
			{
				Name:   "Low",
				Value:  q.low,
				Inline: true,
			},
			{
				Name:   "P/E",
				Value:  q.pe,
				Inline: true,
			},
			{
				Name:   "YTD High",
				Value:  q.yearHigh,
				Inline: true,
			},
			{
				Name:   "YTD Low",
				Value:  q.yearLow,
				Inline: true,
			},
			{
				Name:   "EPS",
				Value:  q.eps,
				Inline: true,
			},
		},
	}
	s.ChannelMessageSendEmbed(channelID, e)
}
