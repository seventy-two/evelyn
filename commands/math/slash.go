package math

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "wa", Description: "Wolfram Alpha", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("query", "What to ask Wolfram Alpha", true)}},
	}
}
