package siege

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "r6", Description: "Rainbow Six Siege player stats", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("username", "Player username", true)}},
	}
}
