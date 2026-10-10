package movie

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "m", Description: "Movie info", Handler: invokeCommand,
			Options: []*discordgo.ApplicationCommandOption{
				slash.StringOpt("title", "Movie title", true),
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "year", Description: "Release year"},
			},
			Text: func(o slash.Options) string {
				if y := o.Int("year"); y != 0 {
					return fmt.Sprintf("!m %s (%d)", o.String("title"), y)
				}
				return "!m " + o.String("title")
			}},
	}
}
