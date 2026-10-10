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
	WordnikSearchURL       string
	WordnikRelationshipURL string
	WOTDURL                string
	WordnikAPIKey          string
}

var serviceConfig *Service

func dict(matches []string) (msg string, err error) {
	text := url.QueryEscape(strings.Join(matches, "+"))
	var data []Wordnik
	var result []string

	err = web.GetJSON(fmt.Sprintf(serviceConfig.WordnikSearchURL, text, serviceConfig.WordnikAPIKey), &data)
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

func synonym(matches []string) (msg string, err error) {
	text := url.QueryEscape(strings.Join(matches, "+"))
	var data []RelatedWords
	var result []string

	err = web.GetJSON(fmt.Sprintf(serviceConfig.WordnikRelationshipURL, text, serviceConfig.WordnikAPIKey), &data)
	if err != nil {
		if strings.Contains(err.Error(), "Not found") {
			return fmt.Sprintf("not found"), nil
		}
		return fmt.Sprintf("There was a problem with your request."), err
	}
	if len(data) == 0 {
		return fmt.Sprintf("Word/phrase not found."), nil
	}
	for _, word := range data {
		for _, related := range word.Words {
			result = append(result, related)
		}
	}

	out := strings.Join(result, ", ")

	return out, nil
}

func wotd() (*discordgo.MessageEmbed, error) {
	data := &wotdResponse{}
	err := web.GetJSON(fmt.Sprintf(serviceConfig.WOTDURL, serviceConfig.WordnikAPIKey), data)
	if err != nil {
		return nil, err
	}
	if data.Word == "" {
		return nil, fmt.Errorf("no word of the day")
	}
	return wotdEmbed(data), nil
}

// wotdEmbed formats the word of the day as a card: up to three definitions,
// an example sentence and the note on the word's origin.
func wotdEmbed(data *wotdResponse) *discordgo.MessageEmbed {
	var defs []string
	for i, d := range data.Definitions {
		if i == 3 {
			break
		}
		def := d.Text
		if d.PartOfSpeech != "" {
			def = fmt.Sprintf("*%s* %s", d.PartOfSpeech, def)
		}
		defs = append(defs, fmt.Sprintf("%d. %s", i+1, def))
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "Word of the day"},
		Title:       data.Word,
		URL:         "https://www.wordnik.com/words/" + url.PathEscape(data.Word),
		Description: truncate(strings.Join(defs, "\n"), 4096),
		Color:       0x5865F2,
	}
	if len(data.Examples) > 0 {
		ex := data.Examples[0]
		text := fmt.Sprintf("> %s", ex.Text)
		if ex.Title != "" {
			text += fmt.Sprintf("\n— *%s*", ex.Title)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Example", Value: truncate(text, 1024)})
	}
	if data.Note != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Origin", Value: truncate(data.Note, 1024)})
	}
	return embed
}

// truncate shortens s to at most n characters, ending with an ellipsis if cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
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
	case "!syn":
		str, err = synonym(matches[1:])
	case "!wotd":
		embed, err := wotd()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```an error occured (%s)```", err))
			return
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	if err != nil {
		str = fmt.Sprintf("an error occured (%s)", err)
	}

	if str != "" {
		fmtstr := fmt.Sprintf("```%s```", str)
		s.ChannelMessageSend(m.ChannelID, fmtstr)
	}
}
