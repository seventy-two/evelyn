package olympics

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/mgutz/ansi"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/columnize"
)

type Service struct {
	MedalURL    string
	EventURL    string
	ScheduleURL string
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

	loc, _ = time.LoadLocation("Europe/London")
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
	sort.Slice(o.Medals, func(i, j int) bool {
		return o.Medals[i].Gold > o.Medals[j].Gold
	})
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

func eventsMobile() (*discordgo.MessageEmbed, error) {
	m := &Events{}
	err := web.GetJSON(serviceConfig.EventURL, m)
	if err != nil {
		return nil, err
	}

	return eventsToStringsMobile(m), nil
}

func schedule() ([]string, error) {
	m := &Schedule{}
	err := web.GetJSON(serviceConfig.ScheduleURL, m)
	if err != nil {
		return nil, err
	}

	return scheduleToStrings(m), nil
}

func eventsToStringsMobile(e *Events) *discordgo.MessageEmbed {
	sort.Slice(e.Events, func(i, j int) bool {
		return e.Events[i].EndTime.After(e.Events[j].EndTime)
	})

	emb := &discordgo.MessageEmbed{
		Title: "Paris 2024",
	}

	for i, ev := range e.Events {
		if i > 5 {
			break
		}

		var s string
		for _, gold := range ev.Gold {
			s = fmt.Sprintf(":first_place: %s (%s)\n", gold.Country, gold.CompetitorName)
		}
		for _, silver := range ev.Silver {
			s = fmt.Sprintf("%s:second_place: %s (%s)\n", s, silver.Country, silver.CompetitorName)
		}
		for _, bronze := range ev.Bronze {
			s = fmt.Sprintf("%s:third_place: %s (%s)\n", s, bronze.Country, bronze.CompetitorName)
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s - %s %s", ev.Discipline, ev.Gender, ev.Name),
			Value:  s,
			Inline: true,
		})

	}

	emb.Footer = &discordgo.MessageEmbedFooter{
		Text: "https://www.youtube.com/watch?v=cZz1oamNbng",
	}

	return emb
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

		s = b(fmt.Sprintf("%s - %s %s", ev.Discipline, ev.Gender, ev.Name))
		for _, gold := range ev.Gold {
			s = fmt.Sprintf("%s|%s", s, g(fmt.Sprintf("%s (%s)", gold.ID, gold.CompetitorName)))
		}
		for _, silver := range ev.Silver {
			s = fmt.Sprintf("%s|%s", s, si(fmt.Sprintf("%s (%s)", silver.ID, silver.CompetitorName)))
		}
		for _, bronze := range ev.Bronze {
			s = fmt.Sprintf("%s|%s", s, br(fmt.Sprintf("%s (%s)", bronze.ID, bronze.CompetitorName)))
		}

		out = append(out, s)
	}
	return out
}

func scheduleToStrings(sch *Schedule) []string {
	var out []string
	for _, day := range *sch {
		t, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			return nil
		}
		if t.Before(time.Now().AddDate(0, 0, -1)) {
			continue
		}

		for _, ev := range day.Events {
			if ev.Status != "Live" {
				continue
			}
			if ev.MedalAwarded == "None" {
				continue
			}

			var timeString string
			switch {
			case IsToday(t):
				timeString = fmt.Sprintf("Today %s", ev.StartTime.Local().Format("15:04"))
			case IsTomorrow(t):
				timeString = fmt.Sprintf("Tomorrow %s", ev.StartTime.Format("15:04"))
			default:
				timeString = fmt.Sprintf("%s %s", t.Format("02/01"), ev.StartTime.Format("15:04"))
			}

			s := fmt.Sprintf("%s | %s | %s - %s %s", ev.Status, timeString, ev.Discipline, ev.Gender, ev.Name)
			if len(ev.Competitors) == 2 {
				s = fmt.Sprintf("%s | %s vs %s", s, ev.Competitors[0].Name, ev.Competitors[1].Name)
			}
			out = append(out, s)
		}
		for _, ev := range day.Events {
			if ev.Status != "Upcoming" {
				continue
			}
			if ev.MedalAwarded == "None" {
				continue
			}
			if len(out) > 10 {
				continue
			}

			var timeString string
			switch {
			case IsToday(t):
				timeString = fmt.Sprintf("Today %s", ev.StartTime.In(loc).Format("15:04"))
			case IsTomorrow(t):
				timeString = fmt.Sprintf("Tomorrow %s", ev.StartTime.In(loc).Format("15:04"))
			default:
				timeString = fmt.Sprintf("%s %s", t.Format("02/01"), ev.StartTime.In(loc).Format("15:04"))
			}

			s := fmt.Sprintf("%s | %s | %s - %s %s", ev.Status, timeString, ev.Discipline, ev.Gender, ev.Name)
			if len(ev.Competitors) == 2 {
				s = fmt.Sprintf("%s | %s vs %s", s, ev.Competitors[0].Name, ev.Competitors[1].Name)
			}
			out = append(out, s)
		}
		if len(out) > 10 {
			continue
		}
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
	case "!mevents":
		res, err := eventsMobile()
		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
			s.ChannelMessageSend(m.ChannelID, str)
		}
		s.ChannelMessageSendEmbed(m.ChannelID, res)

	case "!schedule":
		res, err := schedule()
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

func IsToday(date time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsTomorrow(date time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := time.Now().AddDate(0, 0, 1).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
