package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageRecord(t *testing.T) {
	ts := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	m := &discordgo.MessageCreate{Message: &discordgo.Message{
		ID: "3", GuildID: "g", ChannelID: "c", Content: "hello <@5>", Timestamp: ts, Type: discordgo.MessageTypeReply,
		Author:           &discordgo.User{ID: "1", Username: "seventytwo", GlobalName: "SeventyTwo"},
		Member:           &discordgo.Member{Nick: "72"},
		MessageReference: &discordgo.MessageReference{MessageID: "2"},
	}}
	b, err := json.Marshal(newMessageRecord(m))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"id":"3","guild_id":"g","channel_id":"c","author_id":"1","author_name":"72","type":19,"timestamp":"2026-10-10T12:00:00Z","content":"hello <@5>","reply_to":"2"}`, string(b))

	m.Member = nil
	assert.Equal(t, "SeventyTwo", newMessageRecord(m).AuthorName, "global name when there's no nickname")
	m.Author.GlobalName = ""
	assert.Equal(t, "seventytwo", newMessageRecord(m).AuthorName)
}
