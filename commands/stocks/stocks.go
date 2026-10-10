package stocks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Service struct {
	QuoteURL    string
	MoverURL    string
	ExchangeURL string
	LookupURL   string
	EarningsURL string
	APIKey      string
	Storage     *WatchlistStorage
	LogoURL     string
	LogoAPIKey  string
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

// allowedChannel reports whether stock commands may be used in the channel.
func allowedChannel(channelID string) bool {
	switch channelID {
	case "819577977998147604", "848603390408654878", "612390886785024043":
		return true
	}
	return false
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")

	if !allowedChannel(m.ChannelID) {
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
		outputStock(q, s, m.ChannelID)
	case "!quote":
		var str string
		q, err := lookupAndGetStock(strings.Join(matches[1:], "+"))
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputBigStock(q, s, m.ChannelID)
	case "!watchlist", "!wl":
		serviceConfig.Storage.handleWatchlistRequest(s, m)
	case "!earnings":
		if len(matches) > 1 {
			setEarnings(strings.Join(matches[1:], " "))
		} else if e := getEarnings(); e != "" {
			s.ChannelMessageSend(m.ChannelID, e)
		}
		return
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

	case "!nasdaq", "!s&p", "!snp", "!sp", "!dow":
		var q string
		switch matches[0] {
		case "!nasdaq":
			q = "NASDAQ100"
		case "!s&p", "!snp", "!sp":
			q = "SP500"
		case "!dow":
			q = "DOW30"
		}

		top, bottom, err := GetMovers(q)
		if err != nil {
			str := fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputMovers(top, bottom, s, m.ChannelID)
	}
}
