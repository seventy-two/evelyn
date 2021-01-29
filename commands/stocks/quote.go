package stocks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func getStock(query string) (msg *stock, err error) {
	lookup := &Lookup{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.LookupURL, query), lookup)
	if err != nil {
		return nil, err
	}
	if len(lookup.ResultSet.Result) == 0 {
		return nil, nil
	}
	data := &IEXStocks{}

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
	url := fmt.Sprintf(serviceConfig.QuoteURL, symbol, serviceConfig.APIKey)

	err = web.GetJSON(url, data)
	if err != nil {
		return nil, nil
	}

	if data.CompanyName == "" {
		return nil, nil
	}

	return &stock{
		symbol:        data.Symbol,
		name:          data.CompanyName,
		latestPrice:   data.LatestPrice,
		latestTime:    data.LatestTime,
		latestSource:  data.LatestSource,
		change:        data.Change,
		changePercent: data.ChangePercent,
		week52high:    data.Week52High,
		week52low:     data.Week52Low,
		ytdChange:     data.YtdChange,
		peRatio:       data.PeRatio,
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
	ytdPlus := ""
	if q.ytdChange > 0 {
		ytdPlus = "+"
	}

	s.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
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
	})
}
