package siege

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Service struct {
	AuthURL      string
	AuthUser     string
	AuthPassword string
	AuthToken    string
	AuthExpiry   *time.Time
	ProfileURL   string
	LevelURL     string
	RankURL      string
}

var serviceConfig *Service

func siege(user string) (string, error) {
	client := &http.Client{}

	if serviceConfig.AuthExpiry == nil || serviceConfig.AuthExpiry.Before(time.Now()) {
		tok, expiry, err := authenticate(client, serviceConfig.AuthURL, serviceConfig.AuthUser, serviceConfig.AuthPassword)
		if err != nil {
			return "", err
		}
		serviceConfig.AuthToken = tok
		serviceConfig.AuthExpiry = expiry
	}

	id, err := retrieveUserID(client, user, serviceConfig.ProfileURL, serviceConfig.AuthToken)
	if err != nil {
		return "", err
	}
	profile, err := retrievePlayerProfile(client, id, serviceConfig.LevelURL, serviceConfig.AuthToken)
	if err != nil {
		return "", err
	}
	level := strconv.Itoa(profile.Level)

	player, err := retrievePlayer(client, id, serviceConfig.RankURL, serviceConfig.AuthToken)
	if err != nil {
		return "", err
	}

	rank, mmr, season := player.Rank, strconv.Itoa(int(player.Mmr)), player.Season

	if rank == 0 {
		out := fmt.Sprintf("%s - Level: %s - Pack Chance: %d%s", user, level, profile.LootboxProbability/100, "%")
		return out, nil
	}

	cleanRank := convertRank(rank)
	cleanSeason := convertSeason(season)

	out := fmt.Sprintf("%s - Level %s - Rank %s - %s MMR - %s", user, level, cleanRank, mmr, cleanSeason)
	return out, nil
}

func convertSeason(season int) string {
	if season == 0 {
		return "Season Unknown"
	}
	year := int(season / 4)
	sub := (season % 4)
	if sub == 0 {
		sub = 4
	}
	return fmt.Sprintf("Year %d Season %d", year, sub)
}

func convertRank(rank int) string {
	ranks := []string{
		"Unranked",
		"Copper IV",
		"Copper III",
		"Copper II",
		"Copper I",
		"Bronze IV",
		"Bronze III",
		"Bronze II",
		"Bronze I",
		"Silver IV",
		"Silver III",
		"Silver II",
		"Silver I",
		"Gold IV",
		"Gold III",
		"Gold II",
		"Gold I",
		"Platinum III",
		"Platinum II",
		"Platinum I",
		"Diamond",
	}
	return ranks[rank]
}

func retrievePlayer(client *http.Client, id, rankURL, auth string) (*player, error) {
	body, err := makeUbiRequest(client, fmt.Sprintf(rankURL, id), auth)
	if err != nil {
		return nil, err
	}
	r := &rankResponse{}
	err = json.Unmarshal(body, r)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, player := range r.Players {
		return &player, nil
	}
	return nil, nil
}

type rankResponse struct {
	Players map[string]player `json:"players"`
}

type player struct {
	BoardID             string    `json:"board_id"`
	PastSeasonsAbandons int       `json:"past_seasons_abandons"`
	UpdateTime          time.Time `json:"update_time"`
	SkillMean           float64   `json:"skill_mean"`
	Abandons            int       `json:"abandons"`
	Season              int       `json:"season"`
	Region              string    `json:"region"`
	ProfileID           string    `json:"profile_id"`
	PastSeasonsLosses   int       `json:"past_seasons_losses"`
	MaxMmr              float64   `json:"max_mmr"`
	Mmr                 float64   `json:"mmr"`
	Wins                int       `json:"wins"`
	SkillStdev          float64   `json:"skill_stdev"`
	Rank                int       `json:"rank"`
	Losses              int       `json:"losses"`
	NextRankMmr         float64   `json:"next_rank_mmr"`
	PastSeasonsWins     int       `json:"past_seasons_wins"`
	PreviousRankMmr     float64   `json:"previous_rank_mmr"`
	MaxRank             int       `json:"max_rank"`
}

func retrievePlayerProfile(client *http.Client, id, levelURL, auth string) (*playerProfile, error) {
	body, err := makeUbiRequest(client, fmt.Sprintf(levelURL, id), auth)
	if err != nil {
		return nil, err
	}
	r := &playerProfiles{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return nil, err
	}

	for _, prof := range r.PlayerProfiles {
		return &prof, nil
	}

	return nil, fmt.Errorf("did not find player profile")
}

type playerProfiles struct {
	PlayerProfiles []playerProfile `json:"player_profiles"`
}

type playerProfile struct {
	Xp                 int    `json:"xp"`
	ProfileID          string `json:"profile_id"`
	LootboxProbability int    `json:"lootbox_probability"`
	Level              int    `json:"level"`
}

func retrieveUserID(client *http.Client, user, profURL, auth string) (string, error) {
	body, err := makeUbiRequest(client, fmt.Sprintf(profURL, user), auth)
	if err != nil {
		return "", err
	}
	fmt.Println(string(body))
	r := &userResponse{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return "", err
	}
	for _, prof := range r.Profiles {
		return prof.IDOnPlatform, nil
	}
	return "", fmt.Errorf("did not find userID")
}

type userResponse struct {
	Profiles []struct {
		ProfileID      string `json:"profileId"`
		UserID         string `json:"userId"`
		PlatformType   string `json:"platformType"`
		IDOnPlatform   string `json:"idOnPlatform"`
		NameOnPlatform string `json:"nameOnPlatform"`
	} `json:"profiles"`
}

func authenticate(client *http.Client, authURL, user, pass string) (string, *time.Time, error) {
	req, err := http.NewRequest("POST", authURL, nil)
	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Ubi-AppId", "39baebad-39e5-4552-8c25-2c9b919064e2")

	// eUser := base64.StdEncoding.EncodeToString([]byte(user))
	// ePass := base64.StdEncoding.EncodeToString([]byte(pass))
	req.SetBasicAuth(user, pass)

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	r := &authResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	err = json.Unmarshal(body, r)
	if err != nil {
		return "", nil, err
	}
	if r.Ticket == "" {
		return "", nil, fmt.Errorf("could not authenticate with ubi")
	}
	return r.Ticket, &r.Expiration, nil
}

func makeUbiRequest(client *http.Client, ubiURL, tok string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(ubiURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Ubi-AppId", "39baebad-39e5-4552-8c25-2c9b919064e2")
	req.Header.Set("Authorization", "Ubi_v1 t="+tok)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type authResponse struct {
	PlatformType                  string      `json:"platformType"`
	Ticket                        string      `json:"ticket"`
	TwoFactorAuthenticationTicket interface{} `json:"twoFactorAuthenticationTicket"`
	ProfileID                     string      `json:"profileId"`
	UserID                        string      `json:"userId"`
	NameOnPlatform                string      `json:"nameOnPlatform"`
	Environment                   string      `json:"environment"`
	Expiration                    time.Time   `json:"expiration"`
	SpaceID                       string      `json:"spaceId"`
	ClientIP                      string      `json:"clientIp"`
	ClientIPCountry               string      `json:"clientIpCountry"`
	ServerTime                    time.Time   `json:"serverTime"`
	SessionID                     string      `json:"sessionId"`
	SessionKey                    string      `json:"sessionKey"`
	RememberMeTicket              interface{} `json:"rememberMeTicket"`
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
	case "!r6":
		res, err := siege(matches[1])

		if err != nil {
			str = fmt.Sprintf("an error occured (%s)", err)
		} else {
			str = res
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
