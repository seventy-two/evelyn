package nfl

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // kickoff times use Europe/London regardless of the host's zone data

	"github.com/bwmarrin/discordgo"
	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

// kickoffZone is the time zone kickoff times are shown in.
var kickoffZone, _ = time.LoadLocation("Europe/London")

func nfl() ([]string, error) {
	sb := &scoreboard{}
	if err := web.GetJSON(serviceConfig.TargetURL, sb); err != nil {
		return nil, err
	}
	return gameLines(sb), nil
}

// gameLines formats each game of the scoreboard as a pipe-separated row for
// columnize: away score | away | @ | home | home score | status.
func gameLines(sb *scoreboard) []string {
	var out []string
	for _, e := range sb.Events {
		if len(e.Competitions) == 0 {
			continue
		}
		c := e.Competitions[0]

		var home, away, homeScore, awayScore, homeID string
		for _, t := range c.Competitors {
			if t.HomeAway == "home" {
				home, homeScore, homeID = t.Team.ShortDisplayName, t.Score, t.Team.ID
			} else {
				away, awayScore = t.Team.ShortDisplayName, t.Score
			}
		}

		var status string
		switch c.Status.Type.State {
		case "pre":
			homeScore, awayScore = "", ""
			status = kickoff(e.Date)
			for _, b := range c.Broadcasts {
				if len(b.Names) > 0 {
					status += " " + b.Names[0]
					break
				}
			}
		case "in":
			status = c.Status.Type.ShortDetail
			if sit := c.Situation; sit != nil {
				if sit.Possession == homeID {
					home = ">" + home
				} else if sit.Possession != "" {
					away = ">" + away
				}
				if sit.DownDistanceText != "" {
					status += " " + sit.DownDistanceText
				}
			}
		default:
			status = c.Status.Type.ShortDetail
		}

		out = append(out, strings.Join([]string{awayScore, away, "@", home, homeScore, status}, " | "))
	}
	return out
}

// kickoff formats an ESPN date (e.g. "2026-10-11T17:00Z") as a UK kickoff time.
func kickoff(date string) string {
	t, err := time.Parse("2006-01-02T15:04Z07:00", date)
	if err != nil {
		return date
	}
	return t.In(kickoffZone).Format("Mon 15:04")
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
	var str string

	switch matches[0] {
	case "!nfl":
		res, err := nfl()

		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		} else {
			str = columnize.SimpleFormat(res)
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
