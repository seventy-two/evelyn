package help

import (
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "help", Description: "List Evelyn's commands", Handler: invokeCommand},
	}
}
