package shitpost

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")
var adamRand = 100

// RegisterService will reg shitpost
func RegisterService(dg *discordgo.Session) {
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.ToLower(m.Content) == "sad" {
		s.ChannelMessageSend(m.ChannelID, "sad")
	}

	if hasWord(m.Content, "same") {
		s.ChannelMessageSend(m.ChannelID, "same")
	}

	if hasWord(m.Content, "brexit") {
		s.ChannelMessageSend(m.ChannelID, brexitCountdown())
	}

	if hasWord(m.Content, "flac") {
		s.ChannelMessageSend(m.ChannelID, flac())
	}

	if hasWord(m.Content, "linux") && !strings.Contains(strings.ToLower(m.Content), "gnu") {
		s.ChannelMessageSend(m.ChannelID, rms())
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
