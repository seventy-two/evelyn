package slash

import (
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

func TestHandleDispatchesTranslatedMessage(t *testing.T) {
	var got *discordgo.MessageCreate
	r := NewRouter(nil)
	r.Add(&Command{Name: "w", Options: []*discordgo.ApplicationCommandOption{StringOpt("location", "", false)},
		Handler: func(_ *discordgo.Session, m *discordgo.MessageCreate) { got = m }})

	// InteractionRespond fails without a real session; dispatch should still happen.
	s, _ := discordgo.New("Bot x")
	s.Client.Transport = failTransport{}
	r.handle(s, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "1", Type: discordgo.InteractionApplicationCommand, ChannelID: "c", GuildID: "g",
		Member: &discordgo.Member{User: &discordgo.User{ID: "u"}},
		Data: discordgo.ApplicationCommandInteractionData{Name: "w", Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Name: "location", Type: discordgo.ApplicationCommandOptionString, Value: "london"},
		}},
	}})

	if assert.NotNil(t, got) {
		assert.Equal(t, "!w london", got.Content)
		assert.Equal(t, "u", got.Author.ID)
		assert.Equal(t, "c", got.ChannelID)
	}
}
