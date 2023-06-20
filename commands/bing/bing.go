package bing

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service
var c *Client

type Client struct {
	serviceConfig *service.Service
}

func New(config *service.Service) *Client {
	return &Client{
		serviceConfig: serviceConfig,
	}
}

func RegisterService(dg *discordgo.Session, config *service.Service) {
	serviceConfig = config
	c = New(serviceConfig)
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")
	if m.Author.ID == "261097001301704704" {
		return
	}
	switch matches[0] {
	case "!bing":
		s.ChannelMessageSend(m.ChannelID, c.GetImage(matches[1:]))
	}
}
