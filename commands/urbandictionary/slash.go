package urbandictionary

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "ud", Description: "Urban Dictionary lookup", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("term", "Term to look up", true)}},
	}
}
