package beer

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "beer", Description: "Beer search", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("name", "Beer name", true)}},
	}
}
