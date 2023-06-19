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
		symbol = query
	}
	return getStock(symbol)
}

func lookupStock(query string) (string, error) {
	lookup := &Lookup{}
	err := web.GetJSON(fmt.Sprintf(serviceConfig.LookupURL, query), lookup)
	if err != nil {
		return "", err
	}
	if len(lookup.ResultSet.Result) == 0 {
		return "", nil
	}
	var symbol string

	for _, res := range lookup.ResultSet.Result {
		if !strings.Contains(res.Symbol, ".") {
			symbol = res.Symbol
			break
		}
	}

	if symbol == "" {
		symbol = strings.Split(lookup.ResultSet.Result[0].Symbol, ".")[0]
	}
	return symbol, nil
}

func getStock(symbol string) (msg *stock, err error) {
	data := &IEXStocks{}
	url := fmt.Sprintf(serviceConfig.QuoteURL, symbol, serviceConfig.APIKey)

	err = web.GetJSON(url, data)
	if err != nil {
		return nil, nil
	}

	if data.Symbol == "" {
		return nil, nil
	}

	return &stock{
		symbol:                data.Symbol,
		name:                  data.CompanyName,
		latestPrice:           data.LatestPrice,
		latestTime:            data.LatestTime,
		latestSource:          data.LatestSource,
		change:                data.Change,
		changePercent:         data.ChangePercent,
		week52high:            data.Week52High,
		week52low:             data.Week52Low,
		ytdChange:             data.YtdChange,
		peRatio:               data.PeRatio,
		extendedPrice:         data.ExtendedPrice,
		extendedChangePercent: data.ExtendedChangePercent,
	}, nil
}

func outputStock(q *stock, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}
	plus := ""
	if q.change > 0 {
		plus = "+"
	}
	ePlus := ""
	if q.extendedChangePercent > 0 {
		ePlus = "+"
	}
	eChange := q.extendedPrice - q.latestPrice
	ytdPlus := ""
	if q.ytdChange > 0 {
		ytdPlus = "+"
	}

	e := &discordgo.MessageEmbed{
		Title:       q.symbol + " - " + q.name,
		Description: q.latestSource + " (" + q.latestTime + ")",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Price",
				Value:  fmt.Sprintf("%.2f", q.latestPrice),
				Inline: true,
			},
			{
				Name:   "Change",
				Value:  fmt.Sprintf("%s%.2f (%s%.2f%s)", plus, q.change, plus, q.changePercent*100, "%"),
				Inline: true,
			},
			{
				Name:   "Year to Date Change",
				Value:  fmt.Sprintf("%s%.2f%s", ytdPlus, q.ytdChange*100, "%"),
				Inline: true,
			},
			{
				Name:   "52 Week High",
				Value:  fmt.Sprintf("%.2f", q.week52high),
				Inline: true,
			},
			{
				Name:   "52 Week Low",
				Value:  fmt.Sprintf("%.2f", q.week52low),
				Inline: true,
			},
			{
				Name:   "P/E Ratio",
				Value:  fmt.Sprintf("%.2f", q.peRatio),
				Inline: true,
			},
		},
	}

	if q.extendedChangePercent != 0 {
		e.Fields = append(e.Fields, []*discordgo.MessageEmbedField{
			{
				Name:   "Extended Price",
				Value:  fmt.Sprintf("%.2f", q.extendedPrice),
				Inline: true,
			},
			{
				Name:   "Change",
				Value:  fmt.Sprintf("%s%.2f (%s%.2f%s)", ePlus, eChange, ePlus, q.extendedChangePercent*100, "%"),
				Inline: true,
			},
		}...)
	}
	s.ChannelMessageSendEmbed(channelID, e)
}
