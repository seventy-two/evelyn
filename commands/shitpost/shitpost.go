package shitpost

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")

// RegisterService will reg shitpost
func RegisterService(dg *discordgo.Session) {
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
}

func hasWord(s, match string) bool {
	fields := strings.Fields(s)
	for _, field := range fields {
		if strings.ToLower(alphanum.ReplaceAllString(field, "")) == match {
			return true
		}
	}
	return false
}
