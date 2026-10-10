package quotes

import (
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "qotd", Description: "Quote of the day", Handler: invokeCommand},
	}
}
