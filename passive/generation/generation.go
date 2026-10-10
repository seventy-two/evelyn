package generation

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genai"
	gi "google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

const (
	model           = "gemini-3.8-flash"
	transcribeModel = "gemini-3.5-transcribe"

	// ownerID is the only user who can reset the chat.
	ownerID = "176722208243187712"

	// mainGuild is the only server the AI features run in.
	mainGuild = "237647756049514498"

	// maxMessageLen is Discord's limit on a message's length.
	maxMessageLen = 2000

	systemInstruction = `Be useful, don't be patronising or write anything that can be portrayed as being patronising, be extremely concise, one sentence responses are best where possible, and do not try to be friendly or personable, just useful and soulless. Your language should reflect the fact you're on a discord with men in their 30s, that are probably a bit tired and grumpy. They are british so use uk english.

Several people talk to you in the same chat. Each message starts with the sender's name, like "dk: what time is it". Don't start your replies with a name. A message replying to something else includes that message first, like (Replying to dk: "...").`
)

var gen *Generation
var serviceConfig *Service

type Service struct {
	APIKey        string
	YoutubeURL    string
	YoutubeAPIKey string
}

type Generation struct {
	c       *genai.Client
	threads *threads
}

// RegisterService registers the AI chat, YouTube summary and voice note
// transcription handler.
func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: config.APIKey,
	})
	if err != nil {
		log.Error(err)
		return
	}

	gen = &Generation{
		c:       client,
		threads: newThreads(5000),
	}

	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.GuildID != mainGuild {
		return
	}
	if strings.Contains(m.Content, "!reset") {
		if m.Author.ID != ownerID {
			s.ChannelMessageSend(m.ChannelID, "Only 72 can reset the chat.")
			return
		}
		gen.threads.reset()
		s.ChannelMessageSend(m.ChannelID, "Chat history reset.")
		return
	}

	mentioned := false
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			mentioned = true
			break
		}
	}

	if url := youtubeURL(m.Content); url != "" {
		summariseYoutube(s, m, url, mentioned)
	} else if mentioned {
		chat(s, m)
	}

	for _, attachment := range m.Attachments {
		if strings.Contains(attachment.URL, ".ogg") {
			transcribe(s, m, attachment.URL)
		}
	}
}

// chat replies to a message that mentions the bot. A reply to one of the
// bot's messages continues the conversation that message came from; anything
// else continues the most recent conversation.
func chat(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelTyping(m.ChannelID)

	ref := referencedMessage(s, m)
	prev := gen.threads.latestID()
	quoted := ref
	if ref != nil {
		if id, ok := gen.threads.lookup(ref.ID); ok {
			// The conversation already contains the replied-to message.
			prev, quoted = id, nil
		}
	}

	text := stripMention(m.Content, s.State.User)
	contents := []gi.Content{textContent(chatInput(displayName(m), text, quoted))}
	if url := imageURL(m.Message, ref); url != "" {
		img, err := imageContent(url)
		if err != nil {
			log.Error(err)
		} else {
			contents = append(contents, img)
		}
	}

	id, out, err := ask(prev, contents, gi.ThinkingLevelLow)
	if err != nil && prev != "" {
		// The conversation may have expired; Google keeps them for a limited
		// time. Start afresh, quoting the replied-to message instead.
		log.Errorf("continuing interaction %s: %s", prev, err)
		contents[0] = textContent(chatInput(displayName(m), text, ref))
		id, out, err = ask("", contents, gi.ThinkingLevelLow)
	}
	if err != nil {
		log.Error(err)
		return
	}
	gen.threads.record(id, reply(s, m, out), true)
}

// summariseYoutube replies with a summary of a linked YouTube video. Long
// videos are only summarised when the bot is mentioned.
func summariseYoutube(s *discordgo.Session, m *discordgo.MessageCreate, url string, mentioned bool) {
	y, err := getYouTube(url)
	if err != nil {
		log.Error(err)
		return
	}
	d := y.GetDuration()
	if d == 0 || (d > 20*time.Minute && !mentioned) {
		return
	}

	s.ChannelTyping(m.ChannelID)
	contents := []gi.Content{
		gi.NewContent(gi.VideoContent{URI: genai.Ptr(url)}),
		textContent(`Summarise this video. Only respond with the summary and no flavour text around responding with the summary like "sure here's the summary."`),
	}
	if extra := strings.TrimSpace(stripMention(strings.Replace(m.Content, url, "", 1), s.State.User)); extra != "" {
		contents = append(contents, textContent(extra))
	}

	id, out, err := ask("", contents, gi.ThinkingLevelMedium)
	if err != nil {
		log.Error(err)
		reply(s, m, fmt.Sprintf("nvm lol %s", err))
		return
	}
	if out == "" {
		reply(s, m, "nvm lol")
		return
	}
	// Replies to the summary continue it, but it doesn't become the
	// conversation plain mentions continue.
	gen.threads.record(id, reply(s, m, out), false)
}

// transcribe replies with a transcription of a voice note.
func transcribe(s *discordgo.Session, m *discordgo.MessageCreate, url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()
	b := bytes.NewBuffer(nil)
	if _, err := b.ReadFrom(resp.Body); err != nil {
		log.Error(err)
		return
	}
	id, out, err := transcribeAudio(b.Bytes(), resp.Header.Get("Content-Type"))
	if err != nil {
		log.Error(err)
		return
	}
	if out == "" {
		return
	}
	gen.threads.record(id, reply(s, m, out), false)
}

// transcribeAudio transcribes audio verbatim with Gemini's speech-to-text
// model, returning the interaction's ID and the transcript.
func transcribeAudio(data []byte, mime string) (string, string, error) {
	res, err := gen.c.Interactions.Create(context.Background(), operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(gi.CreateModelInteraction{
			Model: gi.Model(transcribeModel),
			Input: genai.Ptr(gi.NewInteractionsInput([]gi.Content{audioContent(data, mime)})),
		}),
	})
	if err != nil {
		return "", "", err
	}
	if res.Interaction == nil || res.Interaction.ID == nil {
		return "", "", fmt.Errorf("no interaction in response")
	}
	return *res.Interaction.ID, strings.TrimSpace(outputText(res.Interaction)), nil
}

// ask sends contents to Gemini, continuing from the interaction prev if it
// isn't empty, and returns the new interaction's ID and text output.
func ask(prev string, contents []gi.Content, thinking gi.ThinkingLevel) (string, string, error) {
	req := gi.CreateModelInteraction{
		Model:             gi.Model(model),
		SystemInstruction: genai.Ptr(systemInstruction),
		Input:             genai.Ptr(gi.NewInteractionsInput(contents)),
		GenerationConfig: &gi.GenerationConfig{
			MaxOutputTokens: genai.Ptr(4000),
			ThinkingLevel:   &thinking,
		},
		Tools: []gi.Tool{gi.NewTool(gi.GoogleSearch{})},
	}
	if prev != "" {
		req.PreviousInteractionID = &prev
	}

	res, err := gen.c.Interactions.Create(context.Background(), operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(req),
	})
	if err != nil {
		return "", "", err
	}
	if res.Interaction == nil || res.Interaction.ID == nil {
		return "", "", fmt.Errorf("no interaction in response")
	}
	return *res.Interaction.ID, outputText(res.Interaction), nil
}

// outputText returns the text of the interaction's final run of model output
// steps. The Go SDK doesn't fill in Interaction.OutputText.
func outputText(in *gi.Interaction) string {
	var texts []string
	found := false
	for i := len(in.Steps) - 1; i >= 0; i-- {
		out := in.Steps[i].ModelOutputStep
		if out == nil {
			if found {
				break
			}
			continue
		}
		found = true
		var t strings.Builder
		for _, c := range out.Content {
			if c.TextContent != nil {
				t.WriteString(c.TextContent.Text)
			}
		}
		texts = append([]string{t.String()}, texts...)
	}
	return strings.Join(texts, "")
}

// chatInput formats a chat message for Gemini, naming its sender and quoting
// the message it replies to, if any.
func chatInput(name, text string, quoted *discordgo.Message) string {
	input := fmt.Sprintf("%s: %s", name, text)
	if quoted == nil || quoted.Author == nil {
		return input
	}
	q := strings.TrimSpace(quoted.Content)
	if q == "" {
		q = "[no text]"
	}
	return fmt.Sprintf("(Replying to %s: %q)\n%s", authorName(quoted), q, input)
}

// audioContent wraps audio data, assuming Ogg (Discord's voice note format)
// when the type is unknown.
func audioContent(data []byte, mime string) gi.Content {
	if mime == "" {
		mime = string(gi.AudioContentMimeTypeAudioOgg)
	}
	return gi.NewContent(gi.AudioContent{
		Data:     genai.Ptr(base64.StdEncoding.EncodeToString(data)),
		MimeType: genai.Ptr(gi.AudioContentMimeType(mime)),
	})
}

func textContent(text string) gi.Content {
	return gi.NewContent(gi.TextContent{Text: text})
}

// imageContent downloads an image and prepares it for Gemini.
func imageContent(url string) (gi.Content, error) {
	resp, err := http.Get(url)
	if err != nil {
		return gi.Content{}, err
	}
	defer resp.Body.Close()

	img, err := imaging.Decode(resp.Body)
	if err != nil {
		return gi.Content{}, err
	}
	return pngContent(prepareImage(img))
}

// prepareImage shrinks an image to at most 768px on its longest side and
// flattens it onto white; Gemini misreads transparent images.
func prepareImage(img image.Image) image.Image {
	img = imaging.Fit(img, 768, 768, imaging.Lanczos)
	b := img.Bounds()
	return imaging.Overlay(imaging.New(b.Dx(), b.Dy(), color.White), img, image.Point{}, 1)
}

func pngContent(img image.Image) (gi.Content, error) {
	var b bytes.Buffer
	if err := imaging.Encode(&b, img, imaging.PNG); err != nil {
		return gi.Content{}, err
	}
	return gi.NewContent(gi.ImageContent{
		Data:     genai.Ptr(base64.StdEncoding.EncodeToString(b.Bytes())),
		MimeType: gi.ImageContentMimeTypeImagePng.ToPointer(),
	}), nil
}

// reply sends text as a reply to m, split into several messages if it's too
// long for one, and returns the IDs of the messages sent.
func reply(s *discordgo.Session, m *discordgo.MessageCreate, text string) []string {
	var ids []string
	for _, chunk := range splitMessage(text, maxMessageLen) {
		sent, err := s.ChannelMessageSendReply(m.ChannelID, chunk, m.Reference())
		if err != nil {
			log.Error(err)
			break
		}
		ids = append(ids, sent.ID)
	}
	return ids
}

// splitMessage splits text into chunks of at most max characters, preferring
// to break at newlines, then spaces.
func splitMessage(text string, max int) []string {
	var chunks []string
	r := []rune(strings.TrimSpace(text))
	for len(r) > max {
		cut := max
		if i := lastIndex(r[:max], '\n'); i > max/2 {
			cut = i
		} else if i := lastIndex(r[:max], ' '); i > max/2 {
			cut = i
		}
		chunks = append(chunks, strings.TrimSpace(string(r[:cut])))
		r = []rune(strings.TrimSpace(string(r[cut:])))
	}
	if len(r) > 0 {
		chunks = append(chunks, string(r))
	}
	return chunks
}

func lastIndex(r []rune, c rune) int {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == c {
			return i
		}
	}
	return -1
}

// referencedMessage returns the message m replies to, if any.
func referencedMessage(s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.Message {
	if m.ReferencedMessage != nil {
		return m.ReferencedMessage
	}
	if m.MessageReference == nil {
		return nil
	}
	ref, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
	if err != nil {
		log.Error(err)
		return nil
	}
	return ref
}

// imageURL returns the first image in the message, or failing that in the
// message it replies to.
func imageURL(msgs ...*discordgo.Message) string {
	for _, m := range msgs {
		if m == nil {
			continue
		}
		for _, a := range m.Attachments {
			if strings.Contains(a.ContentType, "image") {
				return a.URL
			}
		}
		for _, e := range m.Embeds {
			if e.Type == discordgo.EmbedTypeImage && e.Image != nil {
				return e.Image.URL
			}
		}
	}
	return ""
}

// youtubeURL returns the first YouTube link in the text.
func youtubeURL(text string) string {
	for _, word := range strings.Fields(text) {
		if strings.Contains(word, "youtube.com") || strings.Contains(word, "youtu.be") {
			return word
		}
	}
	return ""
}

func stripMention(text string, bot *discordgo.User) string {
	text = strings.ReplaceAll(text, fmt.Sprintf("<@%s>", bot.ID), "")
	text = strings.ReplaceAll(text, fmt.Sprintf("<@!%s>", bot.ID), "")
	text = strings.TrimPrefix(strings.TrimSpace(text), "@"+bot.Username)
	// Removing a mention from mid-sentence leaves a double space.
	text = strings.ReplaceAll(text, "  ", " ")
	return strings.TrimSpace(text)
}

func displayName(m *discordgo.MessageCreate) string {
	if m.Member != nil && m.Member.Nick != "" {
		return m.Member.Nick
	}
	return authorName(m.Message)
}

func authorName(m *discordgo.Message) string {
	if m.Member != nil && m.Member.Nick != "" {
		return m.Member.Nick
	}
	if m.Author.GlobalName != "" {
		return m.Author.GlobalName
	}
	return m.Author.Username
}
