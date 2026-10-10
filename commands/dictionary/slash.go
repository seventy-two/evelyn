package dictionary

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "dict", Description: "Dictionary definition", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("word", "Word to define", true)}},
		{Name: "syn", Description: "Synonyms", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("word", "Word to find synonyms for", true)}},
		{Name: "wotd", Description: "Word of the day", Handler: invokeCommand},
	}
}
