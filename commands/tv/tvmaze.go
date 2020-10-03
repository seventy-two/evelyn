package tv

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

type showinfo struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Schedule struct {
		Time string   `json:"time"`
		Days []string `json:"days"`
	} `json:"schedule"`
	Network struct {
		Name string `json:"name"`
	} `json:"network"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Previousepisode struct {
			Href string `json:"href"`
		} `json:"previousepisode"`
		Nextepisode struct {
			Href string `json:"href"`
		} `json:"nextepisode"`
	} `json:"_links"`
}

type nextepisode struct {
	Season  int    `json:"season"`
	Number  int    `json:"number"`
	Airdate string `json:"airdate"`
	Airtime string `json:"airtime"`
}

func tvmaze(matches []string) (msg string, err error) {
	results := &showinfo{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.TargetURL, url.QueryEscape(strings.Join(matches, " "))), results)
	if err != nil {
		return "TVmaze\nCould not find show", nil
	}

	if len(results.Links.Nextepisode.Href) != 0 {
		next := &nextepisode{}
		err = web.GetJSON(results.Links.Nextepisode.Href, next)
		if err != nil {
			return "TVmaze\nCould not find show", nil
		}

		if len(results.Schedule.Days) == 0 {
			results.Schedule.Days = []string{"???"}
		}

		output := fmt.Sprintf("TVmaze\n%s\nAirtime: %s %s on %s\nStatus: %s\nNext Ep: S%vE%v at %s %s",
			results.Name,
			results.Schedule.Days[0],
			results.Schedule.Time,
			results.Network.Name,
			results.Status,
			next.Season,
			next.Number,
			next.Airtime,
			next.Airdate,
		)
		return output, nil
	}

	if len(results.Schedule.Days) == 0 {
		output := fmt.Sprintf("TVmaze\n%s\nStatus: %s",
			results.Name,
			results.Status,
		)
		return output, nil
	}

	output := fmt.Sprintf("TVmaze\n%s\nAirtime: %s %s on %s\nStatus: %s",
		results.Name,
		results.Schedule.Days[0],
		results.Schedule.Time,
		results.Network.Name,
		results.Status,
	)
	return output, nil
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
	case "!tv":
		str, err := tvmaze(matches[1:])
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
