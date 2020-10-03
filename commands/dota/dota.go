package dota

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/seventy-two/Cara/web"
)

// Service is configuration for the Dota service
type Service struct {
	DotaLeagueURL  string
	DotaListingURL string
	DotaMatchURL   string
	DotaHeroesURL  string
	APIKey         string
}

var serviceConfig *Service

const timer = "15:04:05"

type match struct {
	league        string
	viewers       int
	clock         string
	roshan        string
	radiant       string
	radiantScore  int
	radiantNet    int
	radiantHeroes []string
	radiantWins   int
	dire          string
	direScore     int
	direNet       int
	direHeroes    []string
	direWins      int
}

func dotamatches(params []string) ([]*match, error) {
	var matches []*match
	data := &LeagueGames{}
	listing := &LeagueListing{}
	getHeroes := &GetHeroes{}

	heroes := false

	if strings.Contains(strings.Join(params, ""), "h") {
		heroes = true
	}

	err := web.GetJSON(fmt.Sprintf(serviceConfig.DotaListingURL), listing)
	if err != nil {
		return nil, err
	}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.DotaLeagueURL, serviceConfig.APIKey), data)
	if err != nil {
		return nil, err
	}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.DotaHeroesURL, serviceConfig.APIKey), getHeroes)
	if err != nil {
		return nil, err
	}

	for _, game := range data.Result.Games {
		if (game.Spectators >= 1000) || (game.LeagueTier == 3 && game.Spectators >= 200) {
			m := &match{
				radiant:      game.RadiantTeam.TeamName,
				dire:         game.DireTeam.TeamName,
				radiantScore: game.Scoreboard.Radiant.Score,
				direScore:    game.Scoreboard.Dire.Score,
				radiantWins:  game.RadiantSeriesWins,
				direWins:     game.DireSeriesWins,
				viewers:      game.Spectators,
				roshan:       "Killed",
			}

			if game.Scoreboard.RoshanRespawnTimer == 0 {
				m.roshan = "Up"
			}

			for _, info := range listing.Infos {
				if game.LeagueID == info.LeagueID {
					m.league = info.Name
					break
				}
			}

			duration := int(game.Scoreboard.Duration)
			m.clock = fmt.Sprintf((time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC).Add(time.Duration(duration) * time.Second)).Format(timer))

			for _, player := range game.Scoreboard.Radiant.Players {
				m.radiantNet += player.NetWorth
			}
			for _, player := range game.Scoreboard.Dire.Players {
				m.direNet += player.NetWorth
			}

			if heroes {
				for _, pick := range game.Scoreboard.Radiant.Picks {
					m.radiantHeroes = append(m.radiantHeroes, strings.Join(getHerofromID(pick.HeroID, getHeroes), " "))
				}
				for _, pick := range game.Scoreboard.Dire.Picks {
					m.direHeroes = append(m.direHeroes, strings.Join(getHerofromID(pick.HeroID, getHeroes), " "))
				}

			}

			if m.dire == "" {
				m.dire = "Dire"
			}
			if m.radiant == "" {
				m.radiant = "Radiant"
			}
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}

	sort.Slice(matches[:], func(i, j int) bool {
		return matches[i].viewers > matches[j].viewers
	})

	return matches, nil
}

func getHerofromID(id int, heroes *GetHeroes) []string {
	if id == 0 {
		return []string{"PICK"}
	}
	out := getShortHero(id)
	if out != nil {
		return out
	}

	for _, hero := range heroes.Result.Heroes {
		if hero.ID == id {
			return []string{hero.LocalizedName}
		}
	}
	return nil
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

	switch matches[0] {
	case "!d2":
		res, err := dotamatches(matches[1:])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("an error occured (%s)", err))
			return
		}

		if res == nil {
			s.ChannelMessageSend(m.ChannelID, "No games")
			return
		}

		for _, topGame := range res {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Title: topGame.league,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Series Score",
						Value:  fmt.Sprintf("%d - %d", topGame.radiantWins, topGame.direWins),
						Inline: true,
					},
					{
						Name:   "Time",
						Value:  topGame.clock,
						Inline: true,
					},
					{
						Name:   "Dota TV Viewers",
						Value:  fmt.Sprintf("%d", topGame.viewers),
						Inline: true,
					},
					{
						Name:   topGame.radiant,
						Value:  fmt.Sprintf("%d kills | %dg", topGame.radiantScore, topGame.radiantNet),
						Inline: true,
					},
					{
						Name:   topGame.dire,
						Value:  fmt.Sprintf("%d kills | %dg", topGame.direScore, topGame.direNet),
						Inline: true,
					},
					{
						Name:   "Roshan",
						Value:  topGame.roshan,
						Inline: true,
					},
				},
			})
		}
	}
}
