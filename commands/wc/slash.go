package wc

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "wc", Description: "World Cup scores and fixtures", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "What to show (default: recent scores and today's fixtures)",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Upcoming fixtures (next 7 days)", Value: "fixtures"},
					{Name: "Live score updates on", Value: "watch on"},
					{Name: "Live score updates off", Value: "watch off"},
				},
			}}},
	}
}
