package nfl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

func nfl() ([]string, error) {
	m := map[string]game{}
	err := web.GetJSON(serviceConfig.TargetURL, &m)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, g := range m {
		if g.Qtr == nil {
			continue
		}
		game := createGameString(&g)
		if game != "" {
			out = append(out, game)
		}
	}
	return out, nil

}

func createGameString(g *game) string {
	home := getTeamName(g.Home.Abbr)
	away := getTeamName(g.Away.Abbr)
	homeScore := strconv.Itoa(g.Home.Score.T)
	awayScore := strconv.Itoa(g.Away.Score.T)
	down := ""
	empty := ""
	clock := ""

	if *g.Qtr != "Pregame" && *g.Qtr != "Final" {
		if g.Posteam == g.Home.Abbr {
			home = ">" + home
		} else {
			away = ">" + away
		}
		switch g.Down {
		case 1:
			down = "1st & " + strconv.Itoa(g.Togo)
		case 2:
			down = "2nd & " + strconv.Itoa(g.Togo)
		case 3:
			down = "3rd & " + strconv.Itoa(g.Togo)
		case 4:
			down = "4th & " + strconv.Itoa(g.Togo)
		}

		if g.Yl != "" {
			down = down + " at " + g.Yl
		}
		clock = g.Clock
	} else if *g.Qtr == "Pregame" {
		return ""
	}

	return awayScore + " | " + away + " | @ | " + home + "  | " + homeScore + empty + " | " + clock + getQuarter(*g.Qtr) + " " + down
}

func getQuarter(q string) string {
	switch q {
	case "1", "2", "3", "4":
		return "Q" + q
	case "final":
		return "F"
	case "final overtime":
		return "F OT"
	default:
		return q
	}
}

func getTeamName(team string) string {
	switch team {
	case "ARI":
		return "Cardinals"
	case "ATL":
		return "Falcons"
	case "CAR":
		return "Panthers"
	case "CHI":
		return "Bears"
	case "DAL":
		return "Cowboys"
	case "DET":
		return "Lions"
	case "GB":
		return "Packers"
	case "MIN":
		return "Vikings"
	case "NO":
		return "Saints"
	case "NYG":
		return "Giants"
	case "PHI":
		return "Eagles"
	case "LAR":
		return "Rams"
	case "SF":
		return "49ers"
	case "SEA":
		return "Seahawks"
	case "TB":
		return "Buccaneers"
	case "WAS":
		return "Redskins"
	case "BAL":
		return "Ravens"
	case "BUF":
		return "Bills"
	case "CIN":
		return "Bengals"
	case "CLE":
		return "Browns"
	case "DEN":
		return "Broncos"
	case "HOU":
		return "Texans"
	case "IND":
		return "Colts"
	case "JAC":
		return "Jaguars"
	case "KC":
		return "Chiefs"
	case "MIA":
		return "Dolphins"
	case "NE":
		return "Patriots"
	case "NYJ":
		return "Jets"
	case "OAK":
		return "Raiders"
	case "PIT":
		return "Steelers"
	case "LAC":
		return "Chargers"
	case "TEN":
		return "Titans"
	default:
		return team
	}
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
