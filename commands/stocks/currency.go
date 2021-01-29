package stocks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func getCurrency(symbol string) (*stock, error) {
	c := &Currency{}
	url := fmt.Sprintf(serviceConfig.CurrencyURL, symbol, serviceConfig.APIKey)
	err := web.GetJSON(url, c)
	if err != nil {
		return nil, err
	}

	if len(c.Results) < 1 {
		return nil, fmt.Errorf("no results")
	}

	data := c.Results[0]

	if data.Symbol == "" {
		return nil, nil
	}

	return &stock{
		symbol:      data.Symbol,
		latestPrice: data.Rate,
	}, nil
}

func outputCurrency(q *stock, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	s.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: q.symbol,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Rate",
				Value:  fmt.Sprintf("%.2f", q.latestPrice),
				Inline: true,
			},
		},
	})
}
