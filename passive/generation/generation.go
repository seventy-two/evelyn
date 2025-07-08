package generation

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
	"github.com/google/generative-ai-go/genai"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var gen *Generation
var serviceConfig *Service

type Service struct {
	APIKey        string
	YoutubeURL    string
	YoutubeAPIKey string
}

type Generation struct {
	c    *genai.Client
	m    *genai.GenerativeModel
	chat *genai.ChatSession
}

func initClient(key string) (*genai.Client, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(key))
	if err != nil {
		log.Error(err)
	}
	return client, err
}

func (g *Generation) resetChat() {
	g.chat = g.m.StartChat()
}

// RegisterService will reg shitpost
func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	client, err := initClient(config.APIKey)
	if err != nil {
		log.Error(err)
		return
	}
	m := client.GenerativeModel("gemini-2.0-flash-lite")
	m.SetMaxOutputTokens(200)
	m.SystemInstruction = genai.NewUserContent(
		genai.Text("be useful, don't be patronising or write anything that can be portrayed as being patronising, be extremely concise, one sentence responses are best where possible, and do not try to be friendly or personable, just useful and soulless"))
	m.SetCandidateCount(1)
	chat := m.StartChat()
	gen = &Generation{
		c:    client,
		m:    m,
		chat: chat,
	}
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			s.ChannelTyping(m.ChannelID)
			if strings.Contains(m.Content, "youtube.com") || strings.Contains(m.Content, "youtu.be") {
				break
			}
			imgUrl := checkForImages(s, m.Message)
			var err error
			str := strings.Trim(m.Content, fmt.Sprintf("<@!%s>", s.State.User.ID))
			if imgUrl != "" {
				str, err = generateTextWithImage(str, imgUrl)
				if err != nil {
					log.Error(err)
					return
				}
			} else {
				str, err = generateText(str)
				if err != nil {
					log.Error(err)
					return
				}
			}
			_, err = s.ChannelMessageSendReply(m.ChannelID, str, m.Reference())
			if err != nil {
				log.Error(err)
				return
			}
			break
		}
	}
	if strings.Contains(m.Content, "youtube.com") || strings.Contains(m.Content, "youtu.be") {
		words := strings.Fields(m.Content)
		var url string
		for _, word := range words {
			if strings.Contains(word, "youtube.com") || strings.Contains(word, "youtu.be") {
				url = word
				break
			}
		}
		if url == "" {
			return
		}
		content := strings.Replace(m.Content, url, "", -1)
		s.ChannelTyping(m.ChannelID)
		y, err := getYouTube(url)
		if err != nil {
			log.Error(err)
			return
		}
		if y.GetDuration() == 0 {
			return
		}
		if y.GetDuration() > 20*time.Minute {
			mentioned := false
			for _, mention := range m.Mentions {
				if mention.ID == s.State.User.ID {
					mentioned = true
					break
				}
			}
			if !mentioned {
				return
			}
		}
		_, err = s.ChannelMessageSendReply(m.ChannelID, "1 sec", m.Reference())
		if err != nil {
			log.Error(err)
			return
		}
		str, err := generateYoutubeSummary(url, content)
		if err != nil {
			log.Error(err)
			s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("nvm lol %s", err.Error()), m.Reference())
			return
		}
		if str == "" {
			s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("nvm lol %s", err.Error()), m.Reference())
			return
		}
		_, err = s.ChannelMessageSendReply(m.ChannelID, str, m.Reference())
		if err != nil {
			log.Error(err)
			return
		}
	}

	for _, attachment := range m.Attachments {
		if strings.Contains(attachment.URL, ".ogg") {
			str, err := generateAudioSummary(attachment.URL)
			if err != nil {
				log.Error(err)
				return
			}
			_, err = s.ChannelMessageSendReply(m.ChannelID, str, m.Reference())
			if err != nil {
				log.Error(err)
				return
			}
		}
	}

}

func generateText(input string) (string, error) {
	resp, err := gen.chat.SendMessage(context.Background(), genai.Text(input))
	if err != nil {
		fmt.Println(err)
		log.Error(err)
		return "", err
	}
	var t string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if txt, ok := p.(genai.Text); ok {
				t += string(txt)
			}
		}
	}
	return t, nil
}

func generateYoutubeSummary(url string, content string) (string, error) {
	resp, err := gen.m.GenerateContent(context.Background(),
		genai.FileData{URI: url},
		genai.Text("Summarise this video. Only respond with the summary and no flavour text around responding with the summary."),
		genai.Text(content))
	if err != nil {
		return "", err
	}

	var t string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if txt, ok := p.(genai.Text); ok {
				t += string(txt)
			}
		}
	}
	return t, nil
}

func generateAudioSummary(url string) (string, error) {
	ctx := context.Background()
	fResp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return "", err
	}
	defer fResp.Body.Close()

	file, err := gen.c.UploadFile(ctx, "audio", fResp.Body, &genai.UploadFileOptions{MIMEType: "audio/ogg"})
	if err != nil {
		fmt.Println(err)
		log.Error(err)
	}
	defer gen.c.DeleteFile(ctx, file.Name)

	resp, err := gen.chat.SendMessage(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Transcribe this audio."))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var t string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if txt, ok := p.(genai.Text); ok {
				t += string(txt)
			}
		}
	}
	return t, nil
}

func generateTextWithImage(input string, url string) (string, error) {
	imageResp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer imageResp.Body.Close()
	ext := strings.ToLower(strings.TrimPrefix(imageResp.Header.Get("Content-Type"), "image/"))

	img, err := imaging.Decode(imageResp.Body)
	if err != nil {
		log.Error(err)
		return "", nil
	}

	maxDim := 768
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width > height {
		newWidth := maxDim
		newHeight := (height * maxDim) / width
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	} else {
		newHeight := maxDim
		newWidth := (width * maxDim) / height
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	var imageBytes bytes.Buffer
	imaging.Encode(&imageBytes, img, imaging.PNG)

	resp, err := gen.chat.SendMessage(context.Background(),
		genai.Text(input),
		genai.ImageData(ext, imageBytes.Bytes()))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var t string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if txt, ok := p.(genai.Text); ok {
				t += string(txt)
			}
		}
	}
	return t, nil
}

func checkForImages(s *discordgo.Session, m *discordgo.Message) string {
	for _, attachment := range m.Attachments {
		if strings.Contains(attachment.ContentType, "image") {
			return attachment.URL
		}
	}
	for _, embed := range m.Embeds {
		if embed.Type == discordgo.EmbedTypeImage {
			return embed.Image.URL
		}
	}
	if m.MessageReference != nil {
		refMsg, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			log.Error(err)
			return ""
		}
		return checkForImages(s, refMsg)
	}
	return ""
}
