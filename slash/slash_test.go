package slash

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func opt(name string, v interface{}) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Value: v}
}

func TestDefaultText(t *testing.T) {
	c := &Command{Name: "wl", Options: []*discordgo.ApplicationCommandOption{
		StringOpt("action", "", false), StringOpt("stock", "", false),
	}}
	assert.Equal(t, "!wl", c.text(Options{}))
	assert.Equal(t, "!wl add apple inc", c.text(Options{"stock": opt("stock", "apple inc"), "action": opt("action", "add")}))

	c = &Command{Name: "set", Trigger: "!set location", Options: []*discordgo.ApplicationCommandOption{StringOpt("location", "", true)}}
	assert.Equal(t, "!set location new york", c.text(Options{"location": opt("location", "new york")}))
}

func TestHandleRepliesWithCommandOutput(t *testing.T) {
	rec := &recorder{}
	s, _ := discordgo.New("Bot x")
	s.Client.Transport = rec

	var got *discordgo.MessageCreate
	r := NewRouter(nil)
	r.Add(&Command{Name: "w", Options: []*discordgo.ApplicationCommandOption{StringOpt("location", "", false)},
		Handler: func(s *discordgo.Session, m *discordgo.MessageCreate) {
			got = m
			s.ChannelMessageSend(m.ChannelID, "first")
			s.ChannelMessageSend(m.ChannelID, "second")
			s.ChannelMessageSend("other", "elsewhere")
		}})
	r.handle(s, interaction("w", &discordgo.ApplicationCommandInteractionDataOption{
		Name: "location", Type: discordgo.ApplicationCommandOptionString, Value: "london",
	}))

	if assert.NotNil(t, got) {
		assert.Equal(t, "!w london", got.Content)
		assert.Equal(t, "u", got.Author.ID)
	}
	assert.Equal(t, []string{
		"POST /api/v9/interactions/1/tok/callback",
		"PATCH /api/v9/webhooks/app/tok/messages/@original",
		"POST /api/v9/webhooks/app/tok",
		"POST /api/v9/channels/other/messages",
	}, rec.requests)
}

func TestMessagesAfterFinishGoToChannel(t *testing.T) {
	rec := &recorder{}
	s, _ := discordgo.New("Bot x")
	s.Client.Transport = rec

	reply := newReplySession(s, interaction("w").Interaction)
	reply.session.ChannelMessageSend("c", "reply")
	assert.True(t, reply.finish(0))
	reply.session.ChannelMessageSend("c", "later")

	assert.Equal(t, []string{
		"PATCH /api/v9/webhooks/app/tok/messages/@original",
		"POST /api/v9/channels/c/messages",
	}, rec.requests)
}

func TestFinishWithoutOutput(t *testing.T) {
	rec := &recorder{}
	s, _ := discordgo.New("Bot x")
	s.Client.Transport = rec

	reply := newReplySession(s, interaction("w").Interaction)
	assert.False(t, reply.finish(0))
}

func interaction(name string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "1", AppID: "app", Token: "tok", Type: discordgo.InteractionApplicationCommand,
		ChannelID: "c", GuildID: "g", Member: &discordgo.Member{User: &discordgo.User{ID: "u"}},
		Data: discordgo.ApplicationCommandInteractionData{Name: name, Options: opts},
	}}
}

// recorder is a fake Discord API that records each request and returns an
// empty message.
type recorder struct {
	mu       sync.Mutex
	requests []string
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	r.requests = append(r.requests, req.Method+" "+req.URL.Path)
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"m1"}`)),
		Request:    req,
	}, nil
}
