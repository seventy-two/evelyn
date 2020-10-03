package stocks

import (
	"fmt"
	"sort"
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
	if err := w.Db.RemoveCuntInfo(id, &database.CuntInfo{Watchlist: []string{s}}); err != nil {
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
	if len(matches) > 1 && matches[1] == "add" {
		stock, err := getStock(strings.Join(matches[2:], "+"))
		if stock == nil || err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not add stock to watchlist as stock could not be found")
			return
		}
		if err := w.addStockToWatchlistForUser(m.Author.ID, stock.symbol); err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		return
	}
	if len(matches) > 1 && matches[1] == "remove" {
		if err := w.removeStockFromWatchlistForUser(m.Author.ID, strings.Join(matches[2:], "+")); err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		return
	}

	queries, err := w.retrieveWatchlistForUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	var stocks []*stock
	for _, query := range queries {
		stock, err := getStock(query)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		stocks = append(stocks, stock)
	}

	output := createWatchlistOutput(stocks)
	if output != "" {
		s.ChannelMessageSend(m.ChannelID, output)
	}
}

func createWatchlistOutput(stocks []*stock) string {
	var lines []string
	sort.Slice(stocks, func(i, j int) bool {
		return stocks[i].changePercent > stocks[j].changePercent
	})
	for _, stock := range stocks {
		plus := ""
		if stock.change > 0 {
			plus = "+"
		}
		outputLine := fmt.Sprintf("%s|%.2f|%s%.2f|(%s%.2f%s)", stock.symbol, stock.latestPrice, plus, stock.change, plus, stock.changePercent*100, "%")
		lines = append(lines, outputLine)
	}
	if len(lines) == 0 {
		return ""
	}

	return fmt.Sprintf("```%s```", columnize.SimpleFormat(lines))
}
