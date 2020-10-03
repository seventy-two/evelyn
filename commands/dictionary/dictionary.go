package dictionary

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

// Service is configuration for the Dictionary service
type Service struct {
	WordnikURL    string
	WOTDURL       string
	WordnikAPIKey string
}

var serviceConfig *Service

func dict(matches []string) (msg string, err error) {
	text := url.QueryEscape(strings.Join(matches, "+"))
	var data []Wordnik
	var result []string

	err = web.GetJSON(fmt.Sprintf(serviceConfig.WordnikURL, text, serviceConfig.WordnikAPIKey), &data)
	if err != nil {
		return fmt.Sprintf("There was a problem with your request."), err
	}
	if len(data) == 0 {
		return fmt.Sprintf("Word/phrase not found."), nil
	}
	cap := len(data) // never >3 because limit=3 in URL
	for i := 0; i < cap; i++ {
		result = append(result, fmt.Sprintf("%s - %s\n%s", data[i].Word, data[i].PartOfSpeech, data[i].Text))
	}
	out := ""
	for _, res := range result {
		out += res
		out += "\n"
	}

	return out, nil
}

func wotd() (msg string, err error) {
	data := &wotdResponse{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.WOTDURL, serviceConfig.WordnikAPIKey), data)
	if err != nil {
		return fmt.Sprintf("There was a problem with your request."), nil
	}
	return fmt.Sprintf("Word of the day: %s\n%s - %s", data.Word, data.Note, data.Definitions[0].Text), nil // I really hate doing [0] but we only want one definition. I hate comments that cause horizontal scroll also.
}

func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")
	var str string
	var err error

	switch matches[0] {
	case "!dict":
		str, err = dict(matches[1:])
	case "!wotd":
		str, err = wotd()
	}

	if err != nil {
		str = fmt.Sprintf("an error occured (%s)", err)
	}

	if str != "" {
		fmtstr := fmt.Sprintf("```%s```", str)
		s.ChannelMessageSend(m.ChannelID, fmtstr)
	}
}
