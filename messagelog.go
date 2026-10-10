package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// messageRecord is one line of the JSON message log, which discstats reads
// to build channel statistics.
type messageRecord struct {
	ID         string    `json:"id"`
	GuildID    string    `json:"guild_id,omitempty"`
	ChannelID  string    `json:"channel_id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Bot        bool      `json:"bot,omitempty"`
	Type       int       `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	Content    string    `json:"content"`
	ReplyTo    string    `json:"reply_to,omitempty"`
}

func newMessageRecord(m *discordgo.MessageCreate) messageRecord {
	r := messageRecord{
		ID:         m.ID,
		GuildID:    m.GuildID,
		ChannelID:  m.ChannelID,
		AuthorID:   m.Author.ID,
		AuthorName: m.Author.Username,
		Bot:        m.Author.Bot,
		Type:       int(m.Type),
		Timestamp:  m.Timestamp,
		Content:    m.Content,
	}
	if m.Author.GlobalName != "" {
		r.AuthorName = m.Author.GlobalName
	}
	if m.Member != nil && m.Member.Nick != "" {
		r.AuthorName = m.Member.Nick
	}
	if m.MessageReference != nil {
		r.ReplyTo = m.MessageReference.MessageID
	}
	return r
}

// createMessageLog returns a handler that appends each message received from
// Discord to path as a line of JSON.
func createMessageLog(path string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("opening message log: %v", err)
	}
	var mu sync.Mutex
	enc := json.NewEncoder(f)
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil {
			return
		}
		r := newMessageRecord(m)
		mu.Lock()
		defer mu.Unlock()
		if err := enc.Encode(r); err != nil {
			log.Errorf("writing message log: %v", err)
		}
	}
}
