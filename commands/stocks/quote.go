package stocks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/commands/bing"
)

func lookupAndGetStock(query string) (*stock, error) {
	symbol, err := lookupStock(query)
	if err != nil {
		fmt.Println(err)
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

func lookupStock(query string) (string, error) {
	lookup := &Lookup{}
	err := web.GetJSON(fmt.Sprintf(serviceConfig.LookupURL, query, serviceConfig.APIKey), lookup)
	if err != nil {
		return "", err
	}
	if len(lookup.BestMatches) == 0 {
		return "", nil
	}
	var symbol string

	if symbol == "" {
		symbol = lookup.BestMatches[0].Symbol
	}
	return symbol, nil
}

func outputStock(q *stock, s *discordgo.Session, channelID string, b *bing.Client) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	img := b.GetThumbnail(fmt.Sprintf("%s+logo", q.symbol))

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

func outputBigStock(q *stock, s *discordgo.Session, channelID string, b *bing.Client) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	img := b.GetThumbnail(fmt.Sprintf("%s+logo", q.symbol))

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
