package olympics

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/mgutz/ansi"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/columnize"
)

type Service struct {
	MedalURL string
	EventURL string
}

var (
	serviceConfig *Service

	b  = ansi.ColorFunc("blue")
	g  = ansi.ColorFunc("yellow")
	si = ansi.ColorFunc("white")
	br = ansi.ColorFunc("cyan")
	p  = ansi.ColorFunc("magenta")

	bu  = ansi.ColorFunc("blue+bu")
	gu  = ansi.ColorFunc("yellow+bu")
	su  = ansi.ColorFunc("white+bu")
	bru = ansi.ColorFunc("cyan+bu")
	pu  = ansi.ColorFunc("magenta+bu")
)

func medals() ([]string, error) {
	m := &Olympics{}
	err := web.GetJSON(serviceConfig.MedalURL, m)
	if err != nil {
		return nil, err
	}

	return olympicToStrings(m), nil
}

func olympicToStrings(o *Olympics) []string {
	var out []string
	str := fmt.Sprintf("%s|%s|%s|%s|%s", b("Country"), g("G"), si("S"), br("B"), p("T"))
	out = append(out, str)
	for i, c := range o.Medals {
		if len(out) > 5 && c.ID != "GBR" {
			continue
		}
		country := strings.Replace(c.Country, "China", "Chinese Beaver with Original Cantonese", -1)
		if c.ID == "GBR" && i > 4 {
			country = strings.Replace(c.Country, "Great Britain", fmt.Sprintf("Great Britain (%s)", humanize.Ordinal(i+1)), -1)
		}
		s := fmt.Sprintf("%s|%s|%s|%s|%s", b(country), g(strconv.Itoa(c.Gold)), si(strconv.Itoa(c.Silver)), br(strconv.Itoa(c.Bronze)), p(strconv.Itoa(c.Total)))
		out = append(out, s)
	}
	return out
}

func events() ([]string, error) {
	m := &Events{}
	err := web.GetJSON(serviceConfig.EventURL, m)
	if err != nil {
		return nil, err
	}

	return eventsToStrings(m), nil
}

func eventsToStrings(e *Events) []string {
	sort.Slice(e.Events, func(i, j int) bool {
		return e.Events[i].EndTime.After(e.Events[j].EndTime)
	})

	var out []string
	for _, ev := range e.Events {
		if len(out) > 5 {
			break
		}
		var s string
		if len(ev.Gold) == 1 && ev.Gold[0].CompetitorName != ev.Gold[0].Country {
			s = fmt.Sprintf("%s|%s|%s|%s", b(fmt.Sprintf("%s - %s %s", ev.Discipline, ev.Gender, ev.Name)),
				g(fmt.Sprintf("%s (%s)", ev.Gold[0].Country, ev.Gold[0].CompetitorName)),
				si(fmt.Sprintf("%s (%s)", ev.Silver[0].Country, ev.Silver[0].CompetitorName)),
				br(fmt.Sprintf("%s (%s)", ev.Bronze[0].Country, ev.Bronze[0].CompetitorName)))
		}
		if len(ev.Gold) > 2 {
			s = fmt.Sprintf("%s|%s|%s|%s", b(ev.Name),
				g(ev.Gold[0].Country),
				si(ev.Silver[0].Country),
				br(ev.Bronze[0].Country))
		}
		out = append(out, s)
	}
	return out
}

func RegisterService(dg *discordgo.Session, config *Service) {
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
	case "!medals":
		res, err := medals()
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		} else {
			str = columnize.SimpleFormat(res)
		}

		str = strings.ReplaceAll(str, b("Country"), bu("Country"))
		str = strings.ReplaceAll(str, g("G"), gu("G"))
		str = strings.ReplaceAll(str, si("S"), su("S"))
		str = strings.ReplaceAll(str, br("B"), bru("B"))
		str = strings.ReplaceAll(str, p("T"), pu("T"))

		if str != "" {
			fmtstr := fmt.Sprintf("```ansi\n%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	case "!events":
		res, err := events()
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		} else {
			str = columnize.SimpleFormat(res)
		}
		if str != "" {
			fmtstr := fmt.Sprintf("```ansi\n%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
