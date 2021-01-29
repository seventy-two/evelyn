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

	// if m.Author.ID == "261097001301704704" && (strings.Contains(m.Content, "http://") || strings.Contains(m.Content, "https://") ||
	// 	(strings.Contains(m.Content, "://") || strings.Contains(m.Content, "www")) ||
	// 	strings.Contains(m.Content, "watch?")) {

	// 	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	// dm, err := s.UserChannelCreate(m.Author.ID)
	// 	// if err != nil {
	// 	// 	fmt.Println(err)
	// 	// 	return
	// 	// }
	// 	s.ChannelMessageSend("686275064831934483", m.Content)
	// }

	if strings.HasPrefix(m.Content, "!upset") {
		go upset(s, m)
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
