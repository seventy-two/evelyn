package stocks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

func getCurrencyRates(s *discordgo.Session, channelID string) error {
	r := &CurrencyRates{}
	err := web.GetJSON(serviceConfig.CurrencyRatesURL, r)
	if err != nil {
		return err
	}
	out := fmt.Sprintf("```JPY: %.2fGBP\nEUR: %.2fGBP\nUSD: %.2fGBP\nAUD: %.2fGBP\nCNY: %.2fGBP\n```", r.Rates.JPY, r.Rates.EUR, r.Rates.USD, r.Rates.AUD, r.Rates.CNY)
	s.ChannelMessageSend(channelID, out)
	return nil
}
