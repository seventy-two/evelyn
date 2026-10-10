package tv

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "tv", Description: "TV show info and next episode", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("show", "Show name", true)}},
	}
}
