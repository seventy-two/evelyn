package generation

import (
	"encoding/json"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gi "google.golang.org/genai/interactions/models/interactions"
)

func TestThreads(t *testing.T) {
	th := newThreads(3)
	assert.Equal(t, "", th.latestID())

	th.record("countries", []string{"m1"}, true)
	th.record("cake", []string{"m2", "m3"}, true)
	th.record("summary", []string{"m4"}, false)

	// A reply to the countries answer continues that conversation, even
	// though the cake answer came later.
	id, ok := th.lookup("m1")
	assert.False(t, ok, "oldest message evicted once over the limit")
	id, ok = th.lookup("m2")
	assert.True(t, ok)
	assert.Equal(t, "cake", id)
	id, _ = th.lookup("m4")
	assert.Equal(t, "summary", id)
	assert.Equal(t, "cake", th.latestID(), "summaries don't become the latest conversation")

	th.reset()
	_, ok = th.lookup("m2")
	assert.False(t, ok)
	assert.Equal(t, "", th.latestID())
}

func TestThreadsReplyToEarlierAnswer(t *testing.T) {
	th := newThreads(100)
	th.record("countries", []string{"m1"}, true)
	th.record("cake", []string{"m2"}, true)

	id, ok := th.lookup("m1")
	assert.True(t, ok)
	assert.Equal(t, "countries", id)
}

func TestOutputText(t *testing.T) {
	in := &gi.Interaction{}
	require.NoError(t, json.Unmarshal([]byte(`{"id":"x","steps":[
		{"type":"user_input","content":[{"type":"text","text":"hi"}]},
		{"type":"thought","signature":"abc"},
		{"type":"google_search_call","arguments":{"queries":["q"]},"id":"1"},
		{"type":"model_output","content":[{"type":"text","text":"Hello "},{"type":"text","text":"there"}]},
		{"type":"model_output","content":[{"type":"text","text":"!"}]}
	]}`), in))
	assert.Equal(t, "Hello there!", outputText(in))
	assert.Equal(t, "", outputText(&gi.Interaction{}))
}

func TestChatInput(t *testing.T) {
	assert.Equal(t, "dk: list 10 countries", chatInput("dk", "list 10 countries", nil))

	quoted := &discordgo.Message{Content: "Kenya, Ghana", Author: &discordgo.User{Username: "evelyn"}}
	assert.Equal(t, "(Replying to evelyn: \"Kenya, Ghana\")\ndk: do it again", chatInput("dk", "do it again", quoted))

	quoted = &discordgo.Message{Author: &discordgo.User{Username: "bob", GlobalName: "Bob"}}
	assert.Equal(t, "(Replying to Bob: \"[no text]\")\ndk: what's this", chatInput("dk", "what's this", quoted))
}

func TestSplitMessage(t *testing.T) {
	assert.Equal(t, []string{"short"}, splitMessage("short", 2000))
	assert.Empty(t, splitMessage("  ", 2000))

	long := strings.Repeat("word ", 1000) // 5000 chars
	chunks := splitMessage(long, 2000)
	assert.Len(t, chunks, 3)
	for _, c := range chunks {
		assert.LessOrEqual(t, len([]rune(c)), 2000)
		assert.False(t, strings.HasSuffix(c, " wor"), "breaks between words")
	}
	assert.Equal(t, strings.TrimSpace(long), strings.Join(chunks, " "))

	lines := strings.Repeat("a", 1500) + "\n" + strings.Repeat("b", 1000)
	assert.Equal(t, []string{strings.Repeat("a", 1500), strings.Repeat("b", 1000)}, splitMessage(lines, 2000))
}

func TestStripMention(t *testing.T) {
	bot := &discordgo.User{ID: "42", Username: "evelyn"}
	assert.Equal(t, "hello", stripMention("<@42> hello", bot))
	assert.Equal(t, "hello there", stripMention("hello <@!42> there", bot))
	assert.Equal(t, "hi", stripMention("@evelyn hi", bot))
}

func TestYoutubeURL(t *testing.T) {
	assert.Equal(t, "https://youtu.be/abc", youtubeURL("look https://youtu.be/abc lol"))
	assert.Equal(t, "", youtubeURL("no links here"))
}

func TestImageURL(t *testing.T) {
	msg := &discordgo.Message{}
	ref := &discordgo.Message{Attachments: []*discordgo.MessageAttachment{{ContentType: "image/png", URL: "ref.png"}}}
	assert.Equal(t, "ref.png", imageURL(msg, ref), "falls back to the replied-to message")
	msg.Attachments = []*discordgo.MessageAttachment{{ContentType: "image/jpeg", URL: "own.jpg"}}
	assert.Equal(t, "own.jpg", imageURL(msg, ref))
	assert.Equal(t, "", imageURL(&discordgo.Message{}, nil))
}

func TestGetDurationNoItems(t *testing.T) {
	assert.Equal(t, 0, int((&Youtube{}).GetDuration()))
}

func TestPrepareImage(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2000, 1000)) // fully transparent
	src.Set(0, 0, color.NRGBA{R: 255, A: 255})

	out := prepareImage(src)
	assert.Equal(t, image.Rect(0, 0, 768, 384), out.Bounds())
	r, g, b, a := out.At(700, 300).RGBA()
	assert.Equal(t, [4]uint32{0xffff, 0xffff, 0xffff, 0xffff}, [4]uint32{r, g, b, a}, "transparency becomes white")

	small := image.NewNRGBA(image.Rect(0, 0, 100, 50))
	assert.Equal(t, image.Rect(0, 0, 100, 50), prepareImage(small).Bounds(), "small images aren't enlarged")
}
