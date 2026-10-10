package stocks

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanuber/columnize"
	"github.com/seventy-two/evelyn/database"
)

type WatchlistStorage struct {
	Db *database.Database
}

func NewWatchlistStorage(db *database.Database) *WatchlistStorage {
	return &WatchlistStorage{Db: db}
}

func (w *WatchlistStorage) addStockToWatchlistForUser(id string, s string) error {
	if err := w.Db.UpdateCuntInfo(id, &database.CuntInfo{Watchlist: []string{s}}); err != nil {
		return err
	}
	return nil
}

func (w *WatchlistStorage) removeStockFromWatchlistForUser(id string, s string) error {
	watchlist, err := w.retrieveWatchlistForUser(id)
	if err != nil {
		return err
	}

	symbol := ""
	for _, entry := range watchlist {
		if entry == s {
			symbol = entry
			break
		}
	}
	if symbol == "" {
		upper := strings.ToUpper(s)
		for _, entry := range watchlist {
			if entry == upper {
				symbol = entry
				break
			}
		}
	}
	if symbol == "" {
		return fmt.Errorf("%s is not on your watchlist", s)
	}

	if err := w.Db.RemoveCuntInfo(id, &database.CuntInfo{Watchlist: []string{symbol}}); err != nil {
		return err
	}
	return nil
}

func (w *WatchlistStorage) retrieveWatchlistForUser(id string) ([]string, error) {
	cunt, err := w.Db.GetCunt(id)
	if err != nil {
		return nil, err
	}
	return cunt.Info.Watchlist, nil
}

func (w *WatchlistStorage) handleWatchlistRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")
	if len(matches) > 1 {
		switch matches[1] {
		case "add":
			s.ChannelTyping(m.ChannelID)
			stock, err := lookupAndGetStock(strings.Join(matches[2:], "+"))
			if stock == nil || err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not add stock to watchlist as stock could not be found")
				return
			}
			if err := w.addStockToWatchlistForUser(m.Author.ID, stock.symbol); err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("added %s to your watchlist", stock.symbol))
			return
		case "remove", "del":
			s.ChannelTyping(m.ChannelID)
			if err := w.removeStockFromWatchlistForUser(m.Author.ID, strings.Join(matches[2:], "+")); err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("removed %s from your watchlist", strings.Join(matches[2:], "+")))
			return
		case "help":
			s.ChannelMessageSend(m.ChannelID, "usage: !wl add/del <stock name> - adds or removes a stock from your watchlist\n!wl - shows the stocks in your watchlist")
			return
		}
	}

	s.ChannelTyping(m.ChannelID)

	queries, err := w.retrieveWatchlistForUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	var stocks []*stock
	query := strings.Join(queries, "|")
	stocks, err = getStockData(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(stocks) > 30 {
		s.ChannelMessageSend(m.ChannelID, "your watchlist is too large, consider deleting some entries :)")
		return
	}

	output := createWatchlistOutput(stocks)
	if output != "" {
		s.ChannelMessageSend(m.ChannelID, output)
	}
}

func createWatchlistOutput(stocks []*stock) string {
	var lines []string
	sort.Slice(stocks, func(i, j int) bool {
		f, _ := strconv.ParseFloat(strings.Trim(stocks[i].percentChange, "%"), 32)
		s, _ := strconv.ParseFloat(strings.Trim(stocks[j].percentChange, "%"), 32)
		return f > s
	})
	for _, stock := range stocks {
		outputLine := fmt.Sprintf("%s|%s|%s|(%s)", stock.symbol, stock.price, stock.change, stock.percentChange)
		lines = append(lines, outputLine)
	}
	if len(lines) == 0 {
		return ""
	}

	c := columnize.SimpleFormat(lines)

	return addAnsiColours(c)
}

var (
	redRe   = regexp.MustCompile(`-[\d.]+%?`)
	greenRe = regexp.MustCompile(`\+[\d.]+%?`)
)

func addAnsiColours(s string) string {
	c := "```ansi\n%s```"
	lines := strings.Split(s, "\n")
	for i, li := range lines {
		if strings.Contains(li, "-") {
			lines[i] = redRe.ReplaceAllStringFunc(li, func(m string) string {
				return "\x1b[2;31m" + m + "\x1b[0m"
			})
		} else if strings.Contains(li, "+") {
			lines[i] = greenRe.ReplaceAllStringFunc(li, func(m string) string {
				return "\x1b[2;32m" + m + "\x1b[0m"
			})
		}
	}
	return fmt.Sprintf(c, strings.Join(lines, "\n"))
}
