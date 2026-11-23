package stocks

import (
	"fmt"
	"strings"

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
		price := s.Last
		change := s.Change
		percentChange := s.ChangePct
		source := s.Source
		time := s.LastTimedate
		extPrice := ""
		extChange := ""
		extPct := ""

		if s.ExtendedMktQuote.LastTime > s.LastTime {
			extPrice = "\n🌙 " + s.ExtendedMktQuote.Last
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
		})
	}
	return stockz, nil
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
		if sym := pickSymbol(query, matches); sym != "" {
			return sym, nil
		}
	}
	return "", lastErr
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

// pickSymbol chooses the symbol to quote from search results: an exact ticker
// match (so "ANT" stays ANT), then the best US stock, then the best US listing
// of any kind, then the best match of all. CNBC ranks the Anthropic ETF (ANTW)
// above ANT for "anthropic", hence preferring stocks over ETFs.
func pickSymbol(query string, matches []LookupMatch) string {
	var found []LookupMatch
	for _, m := range matches {
		if m.Symbol != "" {
			found = append(found, m)
		}
	}
	if len(found) == 0 {
		return ""
	}
	q := strings.ReplaceAll(query, "+", " ")
	for _, m := range found {
		if strings.EqualFold(m.Symbol, q) {
			return m.Symbol
		}
	}
	for _, m := range found {
		if m.CountryCode == "US" && m.IssueType == "STOCK" {
			return m.Symbol
		}
	}
	for _, m := range found {
		if m.CountryCode == "US" {
			return m.Symbol
		}
	}
	return found[0].Symbol
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
