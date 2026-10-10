package help

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

const helpText = "```" + `
Commands
--------
!w [location]            Current weather (uses saved location if omitted)
!f [location]            Weather forecast
!set location <loc>      Save your default location

!dict <word>             Dictionary definition
!syn <word>              Synonyms
!wotd                    Word of the day
!ud <term>               Urban Dictionary lookup

!m <title> [year]        Movie info
!tv <show>               TV show info and next episode

!wa <query>              Wolfram Alpha

!q <symbol>              Quick stock quote
!quote <symbol>          Detailed stock quote
!watchlist / !wl         Manage stock watchlist
!earnings                Upcoming earnings dates
!ex <sym1> <sym2>        Currency / exchange rate
!nasdaq                  NASDAQ 100 top movers
!s&p / !sp               S&P 500 top movers
!dow                     DOW 30 top movers

!nfl                     NFL scores
!d2 [h]                  Dota 2 live league games (add h for heroes)
!r6 <username>           Rainbow Six Siege player stats
!beer <name>             Beer search

!medals                  Olympic medal table
!events                  Olympic events
!schedule                Olympic schedule

!wc                      World Cup scores / fixtures
!wc fixtures             Upcoming WC fixtures (next 7 days)
!wc watch on/off         Live WC score updates

!qotd                    Quote of the day
!reset                   Reset AI chat history

Every command also works as a slash command, e.g. /w, /m, /sp
` + "```"

// RegisterService registers the !help command handler.
func RegisterService(dg *discordgo.Session) {
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.Split(m.Content, " ")[0] == "!help" {
		s.ChannelMessageSend(m.ChannelID, helpText)
	}
}
