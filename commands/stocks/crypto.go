package stocks

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func getCrypto(symbol string) (*stock, error) {
	data := &Crypto{}
	if symbol == "" {
		return nil, nil
	}
	err := web.GetJSON(fmt.Sprintf(serviceConfig.CryptoURL, symbol, serviceConfig.APIKey), data)
	if err != nil {
		return nil, err
	}

	if data.Symbol == "" {
		return nil, nil
	}

	price, err := strconv.ParseFloat(data.LatestPrice, 64)
	if err != nil {
		return nil, err
	}
	return &stock{
		symbol:       data.Symbol,
		latestPrice:  price,
		latestTime:   data.CalculationPrice,
		latestSource: data.LatestSource,
	}, nil
}

func outputCrypto(q *stock, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}

	s.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: q.symbol,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Price",
				Value:  fmt.Sprintf("%.2f", q.latestPrice),
				Inline: true,
			},
			{
				Name:   "Source",
				Value:  fmt.Sprintf("%s", q.latestSource),
				Inline: true,
			},
		},
	})
}
