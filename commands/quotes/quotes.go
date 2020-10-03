package quotes

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

type quotesResult struct {
	Success struct {
		Total int `json:"total"`
	} `json:"success"`
	Contents struct {
		Quotes []struct {
			Quote      string      `json:"quote"`
			Length     interface{} `json:"length"`
			Author     string      `json:"author"`
			Tags       []string    `json:"tags"`
			Category   string      `json:"category"`
			Date       string      `json:"date"`
			Permalink  string      `json:"permalink"`
			Title      string      `json:"title"`
			Background string      `json:"background"`
			ID         string      `json:"id"`
		} `json:"quotes"`
		Copyright string `json:"copyright"`
	} `json:"contents"`
}

type quote struct {
	quote  string
	author string
}

func getQuoteOfTheDay() (*quote, error) {
	results := &quotesResult{}
	err := web.GetJSON(serviceConfig.TargetURL, results)
	if err != nil {
		return nil, err
	}
	if results.Success.Total < 1 {
		return nil, fmt.Errorf("no quote today")
	}
	for _, q := range results.Contents.Quotes {
		return &quote{
			quote:  q.Quote,
			author: q.Author,
		}, nil
	}
	return nil, nil
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
		})
	}
}
