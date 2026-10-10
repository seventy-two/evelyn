package stocks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
)

func GetMovers(exchange string) ([]stock, []stock, error) {
	url := fmt.Sprintf(serviceConfig.MoverURL, exchange)
	m := &Movers{}
	err := web.GetJSON(url, m)
	if err != nil {
		return nil, nil, err
	}
	var top []stock
	var bottom []stock
	for _, list := range m.RankedSymbolList {
		for _, s := range list.RankedSymbols {
			stock := stock{
				symbol:        s.CnbcSymbol,
				name:          s.SymbolDesc,
				price:         fmt.Sprintf("%.2f", s.SymbolData.LastPrice),
				change:        fmt.Sprintf("%.2f", s.SymbolData.Change),
				percentChange: fmt.Sprintf("%.2f%%", s.SymbolData.ChangePct),
			}
			if list.RankOrder == "TOP" {
				top = append(top, stock)
			}
			if list.RankOrder == "BOTTOM" {
				bottom = append(bottom, stock)
			}
		}
	}
	return top, bottom, nil
}

func outputMovers(top []stock, bottom []stock, s *discordgo.Session, channelID string) {
	embed := &discordgo.MessageEmbed{
		Color:  0x7CFC00,
		Fields: []*discordgo.MessageEmbedField{},
	}
	var topDesc []string
	for _, stock := range top {
		name := stock.name
		if len(name) > 30 {
			name = name[:30]
		}
		topDesc = append(topDesc, fmt.Sprintf("%s|%s|+%s|(+%s)", stock.symbol, name, stock.change, stock.percentChange))
	}
	embed.Description = fmt.Sprintf("```%s```", columnize.SimpleFormat(topDesc))
	s.ChannelMessageSendEmbed(channelID, embed)

	embed = &discordgo.MessageEmbed{
		Color:  0xff1919,
		Fields: []*discordgo.MessageEmbedField{},
	}

	bottomDesc := []string{}
	for _, stock := range bottom {
		name := stock.name
		if len(name) > 30 {
			name = name[:30]
		}
		bottomDesc = append(bottomDesc, fmt.Sprintf("%s|%s|%s|(%s)", stock.symbol, name, stock.change, stock.percentChange))
	}
	embed.Description = fmt.Sprintf("```%s```", columnize.SimpleFormat(bottomDesc))
	s.ChannelMessageSendEmbed(channelID, embed)
}
