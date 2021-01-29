package movie

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

func omdb(matches []string, s *discordgo.Session, channelID string) error {
	data := &movie{}
	toQuery := strings.Join(matches, "+")
	err := web.GetJSON(fmt.Sprintf(serviceConfig.TargetURL, toQuery, serviceConfig.APIKey), data)

	if err != nil {
		return fmt.Errorf("There was a problem with your request: %w", err)
	}
	if data.Title == "" {
		return fmt.Errorf("not found")
	}

	out := ""
	for _, r := range data.Ratings {
		out = fmt.Sprintf("%s\n%s", out, fmt.Sprintf("%s: %s", r.Source, r.Value))
	}

	out = strings.Replace(out, "Internet Movie Database", "IMDb", -1)
	out = strings.Replace(out, "Rotten Tomatoes", "RT", -1)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s (%s)", data.Title, data.Year),
		Description: data.Plot,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Genre",
				Value: data.Genre,
			},
			{
				Name:  "Director",
				Value: data.Director,
			},
			{
				Name:  "Actors",
				Value: data.Actors,
			},
		},
	}

	if out != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Ratings",
			Value: out,
		})
	}

	if data.Poster != "N/A" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: data.Poster,
		}
	}

	_, err = s.ChannelMessageSendEmbed(channelID, embed)

	return err
}

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
	case "!m":
		if err := omdb(matches[1:], s, m.ChannelID); err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}
	}
}
