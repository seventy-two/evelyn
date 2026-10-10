package weather

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "w", Description: "Current weather (uses your saved location if omitted)", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("location", "Location", false)}},
		{Name: "f", Description: "Weather forecast (uses your saved location if omitted)", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("location", "Location", false)}},
		{Name: "set", Description: "Save your default weather location", Handler: invokeCommand, Trigger: "!set location",
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("location", "Location", true)}},
	}
}
