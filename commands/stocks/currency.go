package stocks

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func getRate(symbol []string) (*discordgo.MessageEmbed, error) {
	data := &ExchangeRate{}

	err := web.GetJSON(fmt.Sprintf(serviceConfig.ExchangeURL, symbol[0], symbol[1], serviceConfig.APIKey), data)
	if err != nil {
		return nil, err
	}

	if data.RealtimeCurrencyExchangeRate.FromCurrencyCode == "" {
		return nil, nil
	}

	var rate string
	if r, err := strconv.ParseFloat(data.RealtimeCurrencyExchangeRate.ExchangeRate, 32); err == nil {
		rate = fmt.Sprintf("%.2f", r)
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s → %s", data.RealtimeCurrencyExchangeRate.FromCurrencyCode, data.RealtimeCurrencyExchangeRate.ToCurrencyCode),
		Description: fmt.Sprintf("%s to %s", data.RealtimeCurrencyExchangeRate.FromCurrencyName, data.RealtimeCurrencyExchangeRate.ToCurrencyName),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Latest Rate",
				Value:  rate,
				Inline: false,
			},
			{
				Name:   "Last Refresh",
				Value:  data.RealtimeCurrencyExchangeRate.LastRefreshed + " " + data.RealtimeCurrencyExchangeRate.TimeZone,
				Inline: false,
			},
		},
	}, nil
}

func outputExchange(q *discordgo.MessageEmbed, s *discordgo.Session, channelID string) {
	if q == nil {
		s.ChannelMessageSend(channelID, "no results")
		return
	}
	s.ChannelMessageSendEmbed(channelID, q)
}
