package nfl

import (
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "nfl", Description: "NFL scores", Handler: invokeCommand},
	}
}
