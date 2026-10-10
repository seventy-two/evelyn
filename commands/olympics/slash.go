package olympics

import (
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "medals", Description: "Olympic medal table", Handler: invokeCommand},
		{Name: "events", Description: "Olympic events", Handler: invokeCommand},
		{Name: "mevents", Description: "Olympic events (mobile-friendly)", Handler: invokeCommand},
		{Name: "schedule", Description: "Olympic schedule", Handler: invokeCommand},
	}
}
