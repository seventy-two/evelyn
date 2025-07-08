package images

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")

var serviceConfig *Service

type Service struct {
}

// RegisterService will reg images
func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	var img *discordgo.File
	if m.Attachments != nil {
		for _, attachment := range m.Attachments {
			if float32(attachment.Width)/float32(attachment.Height) <= 9.0/16.0 {
				if strings.Contains(attachment.ContentType, "image") {
					img = horizontalImage(attachment.URL)
				}
				if strings.Contains(attachment.ContentType, "video") {
					img = horizontalVideo(attachment.URL)
				}
			}
		}
	}
	if m.Embeds != nil {
		for _, embed := range m.Embeds {
			if embed.Type == discordgo.EmbedTypeImage {
				if embed.Image != nil && (float32(embed.Image.Width)/float32(embed.Image.Height) <= 9.0/16.0) {
					img = horizontalImage(embed.Image.URL)
				}
			}
			if embed.Type == discordgo.EmbedTypeVideo {
				if embed.Video != nil && (float32(embed.Video.Width)/float32(embed.Video.Height) <= 9.0/16.0) {
					img = horizontalVideo(embed.Video.URL)
				}
			}
		}
	}

	if img != nil {
		_, err := s.ChannelFileSendWithMessage(m.ChannelID, "It looks like you posted vertical content :( I've fixed it for you!", fmt.Sprintf(img.Name), img.Reader)
		if err != nil {
			fmt.Println(err)
		}
	}
}
