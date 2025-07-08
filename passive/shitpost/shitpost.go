package shitpost

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/database"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")

var serviceConfig *Service

type Service struct {
	Db *database.Database
}

// RegisterService will reg shitpost
func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.ID == "176722208243187712" {
		if strings.HasPrefix(m.Content, "!cunts") {
			fmt.Println("cunting")
			db := serviceConfig.Db
			c, err := db.GetCunts()
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, cunt := range c {
				fmt.Println(cunt.ID, cunt.Info)
			}
		}
	}
	if m.Author.ID == "261097001301704704" {
		if len(m.Attachments) > 0 || len(m.Embeds) > 0 {
			_, err := s.ChannelMessageSendReply(m.ChannelID,
				"https://media.discordapp.net/attachments/870013827218018346/1208940162407727124/bub.gif?ex=67e00503&is=67deb383&hm=098ea90e9813b5865490ceb9a37030ecfcc6207630e5e0ab6babb1a79bd71764&=&width=400&height=400",
				m.Reference())
			if err != nil {
				fmt.Println(err)
			}
		}
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
