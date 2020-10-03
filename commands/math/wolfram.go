package math

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/evelyn/service"
	"gopkg.in/xmlpath.v2"
)

var serviceConfig *service.Service

func extractURL(text string) string {
	extractedURL := ""
	for _, value := range strings.Split(text, " ") {
		parsedURL, err := url.Parse(value)
		if err != nil {
			continue
		}
		if strings.HasPrefix(parsedURL.Scheme, "http") {
			extractedURL = parsedURL.String()
			break
		}
	}
	return extractedURL
}

func wolfram(matches []string) (msg string, err error) {
	doc, err := http.Get(fmt.Sprintf(serviceConfig.TargetURL, serviceConfig.APIKey, url.QueryEscape(strings.Join(matches, " "))))
	if err != nil {
		return "Wolfram | An error occured", nil
	}
	defer doc.Body.Close()
	root, err := xmlpath.Parse(doc.Body)

	if err != nil {
		return getRandomNumberString(), nil
	}

	success := xmlpath.MustCompile("//queryresult/@success")
	input := xmlpath.MustCompile("//pod[@position='100']//plaintext[1]")
	output := xmlpath.MustCompile("//pod[@position='200']/subpod[1]/plaintext[1]")

	suc, _ := success.String(root)

	if suc != "true" {
		return getRandomNumberString(), nil
	}

	in, _ := input.String(root)
	out, _ := output.String(root)

	in = strings.Replace(in, `\:`, `\n`, -1)
	out = strings.Replace(out, `\:`, `\n`, -1)

	reg := regexp.MustCompile("\\s+")
	in = reg.ReplaceAllString(in, " ")
	out = reg.ReplaceAllString(out, " ")

	in, _ = strconv.Unquote(`"` + in + `"`)
	out, _ = strconv.Unquote(`"` + out + `"`)

	return fmt.Sprintf("Wolfram\n%s >>> %s", in, out), nil
}

func getRandomNumberString() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func RegisterService(dg *discordgo.Session, config *service.Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")

	switch matches[0] {
	case "!wa":
		str, err := wolfram(matches[1:])
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
