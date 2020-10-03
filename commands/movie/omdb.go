package movie

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

func omdb(matches []string) (msg string, err error) {
	data := &movie{}
	toQuery := strings.Join(matches, "+")
	err = web.GetJSON(fmt.Sprintf(serviceConfig.TargetURL, url.QueryEscape(toQuery), serviceConfig.APIKey), data)

	if err != nil {
		return fmt.Sprintf("There was a problem with your request."), nil
	}
	if data.Title == "" {
		return fmt.Sprintf("Not found."), nil
	}
	return fmt.Sprintf("%s (%s)\n%s iMDb: %s\n%s\n%s", data.Title, data.Year, data.Genre, data.ImdbRating, data.Plot, data.Actors), nil
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
		str, err := omdb(matches[1:])
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
