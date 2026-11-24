// Package hearth posts a daily picture of the hearth living room to
// #idiots: how far the room has been lit by the channel's messages, with a
// link to the game. The owner can post one at any time with !hearth.
package hearth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	postedKey   = "hearth:posted"   // the London date of the last daily post
	launchedKey = "hearth:launched" // set once the launch post has gone out
	ownerID     = "176722208243187712"
	// PublicURL is where people play.
	PublicURL = "https://hearth.0x48.io"
)

// Service is the config: the hearth server (HEARTH_URL, local), the
// channel, and when to post each day (UK time).
type Service struct {
	URL       string
	ChannelID string
	Hour      int
	Minute    int
	Store     Store
}

// Store keeps the last post's date (evelyn's Badger DB).
type Store interface {
	GetValue(key string, v interface{}) (bool, error)
	SetValue(key string, v interface{}) error
}

// Status is hearth's /light/status.
type Status struct {
	Count  int       `json:"count"`
	Target int       `json:"target"`
	Start  time.Time `json:"start"`
	Live   bool      `json:"live"`
}

type poster struct {
	config *Service
	client *http.Client
	london *time.Location
	send   func(*discordgo.MessageSend) error
}

// RegisterService starts the daily post and the !hearth command.
func RegisterService(dg *discordgo.Session, config *Service) {
	if config.URL == "" {
		return
	}
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Errorf("hearth: no Europe/London time zone: %s", err)
		return
	}
	p := &poster{
		config: config,
		client: &http.Client{Timeout: 2 * time.Minute}, // rendering takes a few seconds
		london: london,
		send: func(m *discordgo.MessageSend) error {
			m.AllowedMentions = &discordgo.MessageAllowedMentions{}
			_, err := dg.ChannelMessageSendComplex(config.ChannelID, m)
			return err
		},
	}
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == ownerID && m.Content == "!hearth" {
			if err := p.post(time.Now()); err != nil {
				log.Errorf("hearth: posting on request: %s", err)
			}
		}
	})
	go p.run()
}

func (p *poster) run() {
	for {
		if err := p.tick(time.Now()); err != nil {
			log.Errorf("hearth: %s", err)
		}
		time.Sleep(time.Minute)
	}
}

// tick posts the launch announcement once hearth goes live, then a picture
// each day once the post time has passed.
func (p *poster) tick(now time.Time) error {
	var launched bool
	if _, err := p.config.Store.GetValue(launchedKey, &launched); err != nil {
		return fmt.Errorf("loading launch flag: %w", err)
	}
	if !launched {
		var st Status
		if err := p.getJSON("/light/status", &st); err != nil {
			return err
		}
		if !st.Live {
			return nil
		}
		if err := p.post(now); err != nil {
			return err
		}
		if err := p.config.Store.SetValue(launchedKey, true); err != nil {
			return err
		}
		// The launch post counts as today's.
		return p.config.Store.SetValue(postedKey, now.In(p.london).Format("2006-01-02"))
	}
	local := now.In(p.london)
	today := local.Format("2006-01-02")
	due := time.Date(local.Year(), local.Month(), local.Day(), p.config.Hour, p.config.Minute, 0, 0, p.london)
	if local.Before(due) {
		return nil
	}
	var last string
	if _, err := p.config.Store.GetValue(postedKey, &last); err != nil {
		return fmt.Errorf("loading last post: %w", err)
	}
	if last == today {
		return nil
	}
	// Don't post a late catch-up in the evening after downtime.
	if local.Sub(due) > 3*time.Hour {
		return p.config.Store.SetValue(postedKey, today)
	}
	if err := p.post(now); err != nil {
		return err // try again next minute
	}
	return p.config.Store.SetValue(postedKey, today)
}

// post sends the picture and a line about how lit the room is.
func (p *poster) post(now time.Time) error {
	var st Status
	if err := p.getJSON("/light/status", &st); err != nil {
		return err
	}
	if !st.Live {
		return nil // not shared yet
	}
	img, err := p.get("/snapshot.png")
	if err != nil {
		return err
	}
	if err := p.send(&discordgo.MessageSend{
		Content: Caption(st, now.In(p.london)),
		Files:   []*discordgo.File{{Name: "hearth.png", ContentType: "image/png", Reader: bytes.NewReader(img)}},
	}); err != nil {
		return err
	}
	log.Infof("hearth: posted the room at %d messages", st.Count)
	return nil
}

// Caption is the message posted with the picture.
func Caption(st Status, now time.Time) string {
	if st.Count == 0 {
		return "🕯️ **hearth**: the living room is dark. Every message in here lets in a little light, " +
			"until it's fully lit. Until then you're a tiny toy with a sniper rifle, hiding in the shadows: <" + PublicURL + ">"
	}
	pct := min(100, st.Count*100/max(1, st.Target))
	day := int(now.Sub(st.Start).Hours()/24) + 1
	return fmt.Sprintf("🕯️ **hearth, day %d**: %s messages have let in some light, and the room is about %d%% lit. "+
		"Every message makes it a little brighter. Come and hide in it: <%s>", day, thousands(st.Count), pct, PublicURL)
}

func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func (p *poster) get(path string) ([]byte, error) {
	resp, err := p.client.Get(p.config.URL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", path, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 20<<20))
}

func (p *poster) getJSON(path string, v any) error {
	b, err := p.get(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
