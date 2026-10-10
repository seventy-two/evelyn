package quotes

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

// zenQuote is one quote from the ZenQuotes API, which returns a list.
type zenQuote struct {
	Quote  string `json:"q"`
	Author string `json:"a"`
}

type quote struct {
	quote  string
	author string
}

func getQuoteOfTheDay() (*quote, error) {
	var results []zenQuote
	err := web.GetJSON(serviceConfig.TargetURL, &results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0].Quote == "" {
		return nil, fmt.Errorf("no quote today")
	}
	return &quote{
		quote:  results[0].Quote,
		author: results[0].Author,
	}, nil
}

// RegisterService registers the quotes service
func RegisterService(dg *discordgo.Session, config *service.Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")

	switch matches[0] {
	case "!qotd":
		q, err := getQuoteOfTheDay()
		if err != nil || q == nil {
			s.ChannelMessageSend(m.ChannelID, "No quote today")
			return
		}
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       q.author,
			Description: q.quote,
			// ZenQuotes' free tier asks for attribution.
			Footer: &discordgo.MessageEmbedFooter{Text: "Quotes provided by ZenQuotes.io"},
		})
	}
}
