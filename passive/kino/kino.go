// Package kino watches a kino film diary (a small Letterboxd-like site)
// and posts newly logged films, new members and cinema showings to Discord.
package kino

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	cursorKey  = "kino:feed"
	membersKey = "kino:members"
	// collapseOver is how many logs one member can post in a single check
	// before they're summed up in one message instead (say, a Letterboxd
	// sync bringing in a batch).
	collapseOver = 3
	embedColour  = 0xF5B942
)

// Service is the configuration for the kino watcher. It does nothing unless
// URL is set (kino's address, e.g. http://127.0.0.1:6100).
type Service struct {
	URL string
	// Key is kino's bot key (KINO_BOT_KEY there), sent as X-Kino-Key:
	// kino's activity and cinema news are for members only.
	Key       string
	ChannelID string
	Interval  time.Duration
	Store     Store
}

// Store keeps the feed cursor and the members already seen.
type Store interface {
	GetValue(key string, v interface{}) (bool, error)
	SetValue(key string, v interface{}) error
}

// RegisterService starts watching kino in the background.
func RegisterService(dg *discordgo.Session, config *Service) {
	if config.URL == "" {
		return
	}
	if !validURL(config.URL) {
		log.Errorf("kino watcher off: KINO_URL must be an http(s) address, got %q", config.URL)
		return
	}
	w := &watcher{
		config: config,
		client: &http.Client{Timeout: 15 * time.Second},
		send: func(m *discordgo.MessageSend) error {
			m.AllowedMentions = &discordgo.MessageAllowedMentions{} // never ping anyone
			_, err := dg.ChannelMessageSendComplex(config.ChannelID, m)
			return err
		},
		post: func(content string) (string, error) {
			m, err := dg.ChannelMessageSendComplex(config.ChannelID, &discordgo.MessageSend{
				Content: content, AllowedMentions: &discordgo.MessageAllowedMentions{}})
			if err != nil {
				return "", err
			}
			return m.ID, nil
		},
		remove: func(id string) error { return dg.ChannelMessageDelete(config.ChannelID, id) },
	}
	go w.run()
}

type watcher struct {
	config *Service
	client *http.Client
	send   func(*discordgo.MessageSend) error
	post   func(content string) (id string, err error) // for cinema news, kept to delete later
	remove func(id string) error
}

func (w *watcher) run() {
	for {
		if err := w.checkMembers(); err != nil {
			log.Errorf("kino members: %s", err)
		}
		if err := w.checkFeed(); err != nil {
			log.Errorf("kino feed: %s", err)
		}
		if err := w.checkCinema(); err != nil {
			log.Errorf("kino cinema: %s", err)
		}
		time.Sleep(w.config.Interval)
	}
}

func (w *watcher) get(path string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(w.config.URL, "/")+"/api/"+path, nil)
	if err != nil {
		return err
	}
	if w.config.Key != "" {
		req.Header.Set("X-Kino-Key", w.config.Key)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

type user struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

type film struct {
	Label  string `json:"label"`
	Poster string `json:"poster"`
	URL    string `json:"url"`
}

type logEntry struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	User        user   `json:"user"`
	Film        film   `json:"film"`
	WatchedText string `json:"watchedText"`
	Stars       string `json:"stars"`
	Liked       bool   `json:"liked"`
	Rewatch     bool   `json:"rewatch"`
	Review      string `json:"review"`
	Spoilers    bool   `json:"spoilers"`
	Verb        string `json:"verb"`
}

// checkFeed posts logs newer than the stored cursor. The first check only
// records where the feed is.
func (w *watcher) checkFeed() error {
	var cursor int64
	known, err := w.config.Store.GetValue(cursorKey, &cursor)
	if err != nil {
		return fmt.Errorf("loading cursor: %w", err)
	}
	if !known {
		var start struct {
			Latest int64 `json:"latest"`
		}
		if err := w.get("feed?limit=1", &start); err != nil {
			return err
		}
		log.Infof("kino: watching the feed from log %d", start.Latest)
		return w.config.Store.SetValue(cursorKey, start.Latest)
	}
	var page struct {
		Items  []logEntry `json:"items"`
		Latest int64      `json:"latest"`
	}
	if err := w.get(fmt.Sprintf("feed?limit=50&since=%d", cursor), &page); err != nil {
		return err
	}
	if len(page.Items) == 0 {
		return nil
	}
	byUser := map[string][]logEntry{}
	var order []string
	for _, l := range page.Items {
		if _, ok := byUser[l.User.Username]; !ok {
			order = append(order, l.User.Username)
		}
		byUser[l.User.Username] = append(byUser[l.User.Username], l)
	}
	for _, name := range order {
		logs := byUser[name]
		var msgs []*discordgo.MessageSend
		if len(logs) > collapseOver {
			msgs = append(msgs, summary(logs))
		} else {
			for _, l := range logs {
				msgs = append(msgs, &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{l.embed()}})
			}
		}
		for _, m := range msgs {
			if err := w.send(m); err != nil {
				// Keep the cursor so the next check tries again; an
				// earlier member's logs may then post twice, which is
				// better than losing these.
				return fmt.Errorf("posting: %w", err)
			}
		}
	}
	return w.config.Store.SetValue(cursorKey, page.Latest)
}

// checkMembers posts members not seen before. The first check only records
// who's there.
func (w *watcher) checkMembers() error {
	var seen []string
	known, err := w.config.Store.GetValue(membersKey, &seen)
	if err != nil {
		return fmt.Errorf("loading members: %w", err)
	}
	var list struct {
		Members []user `json:"members"`
	}
	if err := w.get("users", &list); err != nil {
		return err
	}
	isSeen := map[string]bool{}
	for _, u := range seen {
		isSeen[u] = true
	}
	changed := !known
	for _, u := range list.Members {
		if isSeen[u.Username] {
			continue
		}
		if known {
			if err := w.send(&discordgo.MessageSend{Content: joinedText(u)}); err != nil {
				log.Errorf("kino: posting new member %s: %s", u.Username, err)
				continue
			}
		}
		isSeen[u.Username] = true
		seen = append(seen, u.Username)
		changed = true
	}
	if !changed {
		return nil
	}
	return w.config.Store.SetValue(membersKey, seen)
}

// markdownEscape stops names and titles turning into formatting.
var markdownEscape = strings.NewReplacer(`\`, `\\`, `*`, `\*`, `_`, `\_`, `~`, `\~`, "`", "\\`", `|`, `\|`, `[`, `\[`, `]`, `\]`, `>`, `\>`)

func joinedText(u user) string {
	name := "**" + markdownEscape.Replace(u.Name) + "**"
	if u.Name != u.Username {
		name += " (" + markdownEscape.Replace(u.Username) + ")"
	}
	return fmt.Sprintf("🎟️ %s just joined kino: <%s>", name, u.URL)
}

func (l logEntry) embed() *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{Name: l.User.Name + " " + l.Verb, URL: l.User.URL},
		Title:  l.Film.Label,
		URL:    l.URL,
		Color:  embedColour,
		Footer: &discordgo.MessageEmbedFooter{Text: "kino"},
	}
	var line []string
	if l.Stars != "" {
		line = append(line, l.Stars)
	}
	if l.Liked {
		line = append(line, "♥")
	}
	if l.Rewatch && l.Verb != "rewatched" {
		line = append(line, "↻ rewatch")
	}
	if l.WatchedText != "" {
		line = append(line, "watched "+l.WatchedText)
	}
	desc := strings.Join(line, "  ·  ")
	if r := strings.TrimSpace(l.Review); r != "" {
		r = excerpt(r, 600)
		if l.Spoilers {
			r = "||" + strings.ReplaceAll(r, "||", "") + "||"
		}
		desc += "\n\n" + r
	}
	e.Description = desc
	if l.Film.Poster != "" {
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: l.Film.Poster}
	}
	return e
}

// summary sums up a batch of one member's logs in one message.
func summary(logs []logEntry) *discordgo.MessageSend {
	u := logs[0].User
	var lines []string
	for i, l := range logs {
		if i == 10 {
			lines = append(lines, fmt.Sprintf("…and %d more", len(logs)-10))
			break
		}
		s := "• [" + markdownEscape.Replace(l.Film.Label) + "](<" + l.URL + ">)"
		if l.Stars != "" {
			s += " " + l.Stars
		}
		lines = append(lines, s)
	}
	return &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{{
		Author:      &discordgo.MessageEmbedAuthor{Name: fmt.Sprintf("%s logged %d films", u.Name, len(logs)), URL: u.URL},
		Description: strings.Join(lines, "\n"),
		Color:       embedColour,
		Footer:      &discordgo.MessageEmbedFooter{Text: "kino"},
	}}}
}

// excerpt cuts text to about n characters at a word break.
func excerpt(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	cut := string([]rune(s)[:n])
	if i := strings.LastIndexAny(cut, " \n"); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " \n.,;:") + "…"
}

// validURL reports whether s is an http(s) address.
func validURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
