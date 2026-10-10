package stocks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/slash"
)

// SlashCommands returns the slash command equivalents of this package's ! commands.
func SlashCommands() []*slash.Command {
	return []*slash.Command{
		{Name: "q", Description: "Quick stock quote", Handler: invokeCommand, Allowed: allowedChannel,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("stock", "Symbol or company name", true)}},
		{Name: "quote", Description: "Detailed stock quote", Handler: invokeCommand, Allowed: allowedChannel,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("stock", "Symbol or company name", true)}},
		{Name: "wl", Description: "Show or manage your stock watchlist", Handler: invokeCommand, Allowed: allowedChannel,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "action",
					Description: "What to do (default: show your watchlist)",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "add", Value: "add"},
						{Name: "remove", Value: "del"},
						{Name: "help", Value: "help"},
					},
				},
				slash.StringOpt("stock", "Stock to add or remove", false),
			}},
		{Name: "earnings", Description: "Upcoming earnings dates", Handler: invokeCommand, Allowed: allowedChannel,
			Options: []*discordgo.ApplicationCommandOption{slash.StringOpt("set", "Replace the earnings text", false)}},
		{Name: "ex", Description: "Currency / exchange rate", Handler: invokeCommand, Allowed: allowedChannel,
			Options: []*discordgo.ApplicationCommandOption{
				slash.StringOpt("from", "Symbol to convert from", true),
				slash.StringOpt("to", "Symbol to convert to", true),
			}},
		{Name: "nasdaq", Description: "NASDAQ 100 top movers", Handler: invokeCommand, Allowed: allowedChannel},
		{Name: "sp", Description: "S&P 500 top movers", Handler: invokeCommand, Allowed: allowedChannel},
		{Name: "dow", Description: "DOW 30 top movers", Handler: invokeCommand, Allowed: allowedChannel},
	}
}
