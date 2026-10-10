package dota

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "d2", Description: "Dota 2 live league games", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "heroes",
				Description: "Show heroes",
			}},
			Text: func(o slash.Options) string {
				if o.Bool("heroes") {
					return "!d2 h"
				}
				return "!d2"
			}},
	}
}
