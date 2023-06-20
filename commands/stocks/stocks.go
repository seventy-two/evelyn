package stocks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/commands/bing"
)

type Service struct {
	QuoteURL    string
	ExchangeURL string
	LookupURL   string
	EarningsURL string
	APIKey      string
	Storage     *WatchlistStorage
	Bing        *bing.Client
}

var serviceConfig *Service

type stock struct {
	symbol           string
	name             string
	price            string
	high             string
	low              string
	change           string
	percentChange    string
	extPrice         string
	extChange        string
	extPercentChange string
	source           string
	eps              string
	pe               string
	mktcap           string
	time             string
	yearHigh         string
	yearLow          string
}

func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")

	if m.ChannelID != "819577977998147604" && m.ChannelID != "848603390408654878" && m.ChannelID != "612390886785024043" {
		return
	}

	switch matches[0] {
	case "!q":
		var str string
		q, err := lookupAndGetStock(strings.Join(matches[1:], "+"))
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputStock(q, s, m.ChannelID, serviceConfig.Bing)
	case "!quote":
		var str string
		q, err := lookupAndGetStock(strings.Join(matches[1:], "+"))
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputBigStock(q, s, m.ChannelID, serviceConfig.Bing)
	case "!watchlist", "!wl":
		serviceConfig.Storage.handleWatchlistRequest(s, m)
	case "!earnings":
		cal := GetCalendar()
		if cal != "" {
			s.ChannelMessageSend(m.ChannelID, cal)
		}
	case "!ex":
		if len(matches) != 3 {
			s.ChannelMessageSend(m.ChannelID, "!ex expects exactly 2 symbols.")
			return
		}
		q, err := getRate(matches[1:])
		if err != nil {
			str := fmt.Sprintf("an error occured - it's likely the symbol is invalid (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputExchange(q, s, m.ChannelID)

	}
}
