package stocks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Service struct {
	QuoteURL         string
	CryptoURL        string
	CurrencyURL      string
	CurrencyRatesURL string
	LookupURL        string
	ChartURL         string
	APIKey           string
	Storage          *WatchlistStorage
}

var serviceConfig *Service

type stock struct {
	symbol                string
	name                  string
	latestPrice           float64
	latestSource          string
	latestTime            string
	change                float64
	changePercent         float64
	week52high            float64
	week52low             float64
	ytdChange             float64
	peRatio               float64
	extendedPrice         float64
	extendedChangePercent float64
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

	if m.ChannelID != "819577977998147604" && m.ChannelID != "848603390408654878" {
		return
	}

	switch matches[0] {
	case "!quote":
		var str string
		q, err := lookupAndGetStock(strings.Join(matches[1:], "+"))
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputStock(q, s, m.ChannelID)
	case "!watchlist", "!wl":
		serviceConfig.Storage.handleWatchlistRequest(s, m)
	case "!crypto":
		if len(matches) < 2 {
			s.ChannelMessageSend(m.ChannelID, "!crypto expects a symbol. Try !cryptos for common crypto prices.")
			return
		}
		q, err := getCrypto(matches[1])
		if err != nil {
			str := fmt.Sprintf("an error occured - it's likely the symbol is invalid (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
		outputCrypto(q, s, m.ChannelID)

	case "!cryptos":
		var c = []string{
			"btcusd",
			"ethusd",
		}

		var out []string
		for _, symbol := range c {
			q, err := getCrypto(symbol)
			if err != nil {
				str := fmt.Sprintf("an error occured (%s)", err)
				s.ChannelMessageSend(m.ChannelID, str)
				return
			}
			out = append(out, fmt.Sprintf("%s: $%.2f", q.symbol, q.latestPrice))
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s```", strings.Join(out, "\n")))
	// case "!currency":
	// 	q, err := getCurrency(matches[1])
	// 	if err != nil {
	// 		str := fmt.Sprintf("an error occured (%s)", err)
	// 		s.ChannelMessageSend(m.ChannelID, str)
	// 		return
	// 	}
	// 	outputCurrency(q, s, m.ChannelID)
	//  PAID USERS ONLY???
	case "!rates":
		err := getCurrencyRates(s, m.ChannelID)
		if err != nil {
			str := fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
			return
		}
	case "!earnings":
		if len(matches) > 1 && userIsGN(s, m) {
			setEarnings(strings.Join(matches[1:], " "))
		} else {
			s.ChannelMessageSend(m.ChannelID, getEarnings())
		}
		return
	case "!chart":
		if len(matches) > 1 {
			makeChart(s, m, matches[1])
		}
	}
}

func userIsGN(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return true
}
