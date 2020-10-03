package urbandictionary

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

type definitionResults struct {
	Tags       []string `json:"tags"`
	ResultType string   `json:"result_type"`
	List       []struct {
		Defid       int    `json:"defid"`
		Word        string `json:"word"`
		Author      string `json:"author"`
		Permalink   string `json:"permalink"`
		Definition  string `json:"definition"`
		Example     string `json:"example"`
		ThumbsUp    int    `json:"thumbs_up"`
		ThumbsDown  int    `json:"thumbs_down"`
		CurrentVote string `json:"current_vote"`
	} `json:"list"`
	Sounds []interface{} `json:"sounds"`
}

func urban(matches []string) (msg string, err error) {
	query := strings.Join(matches, " ")

	results := &definitionResults{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.TargetURL, url.QueryEscape(query)), results)
	if err != nil {
		return fmt.Sprintf("%s | (No definition found)", query), nil
	}
	if results.ResultType == "no_results" || len(results.List) == 0 {
		return fmt.Sprintf("%s | (No definition found)", query), nil
	}

	word := results.List[0].Word
	definition := results.List[0].Definition

	reg := regexp.MustCompile("\\s+")
	definition = reg.ReplaceAllString(definition, " ") // Strip tabs and newlines

	if len(definition) > 480 {
		definition = fmt.Sprintf("%s...", definition[0:480])
	}

	output := fmt.Sprintf("\n%s\n%s", strings.ToTitle(word), definition)

	return output, nil
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
	case "!ud":
		str, err := urban(matches[1:])
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
