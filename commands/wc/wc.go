package wc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
)

// Service holds configuration for the wc command.
type Service struct {
	TargetURL string
}

// actionDetail holds the detail of a single in-game action (goal, card, etc.)
type actionDetail struct {
	key        string // dedup key
	playerName string
	actionType string
	timeLabel  string
	teamName   string
}

// matchSnapshot captures the state of a match at a point in time.
type matchSnapshot struct {
	homeScore   string
	awayScore   string
	status      string
	periodLabel string
	seenKeys    map[string]bool
	actions     []actionDetail
}

// watchSession tracks active watch mode for a channel.
type watchSession struct {
	stop               chan struct{}
	snapshot           map[string]matchSnapshot  // event URN -> snapshot
	gameMessages       map[string]string         // event URN -> last Discord message ID
	pendingGoalActions map[string][]actionDetail // goal actions seen but score not yet confirmed
	pendingGoalKeys    map[string]map[string]bool
}

var (
	svcConfig *Service
	watchMu   sync.Mutex
	watches   = map[string]*watchSession{}
)

// teamColors maps BBC API team names to ANSI colour codes.
// Only Discord-supported colours are used: 30m (dark grey), 31m (red),
// 32m (green), 33m (yellow), 34m (blue), 35m (magenta), 36m (cyan), 37m (white).
// Both FullName and ShortName variants from the BBC API are included.
var teamColors = map[string]string{
	// Group A
	"Mexico":         "\x1b[32m", // green
	"South Korea":    "\x1b[31m", // red
	"Czech Republic": "\x1b[34m", // blue  (BBC FullName/ShortName)
	"Czechia":        "\x1b[34m", // blue  (alternate)
	"South Africa":   "\x1b[33m", // yellow/gold
	// Group B
	"Canada":               "\x1b[31m", // red (maple leaf)
	"Switzerland":          "\x1b[31m", // red
	"Qatar":                "\x1b[35m", // magenta (maroon flag)
	"Bosnia":               "\x1b[34m", // blue (BBC ShortName)
	"Bosnia-Herzegovina":   "\x1b[34m", // blue (BBC FullName)
	"Bosnia & Herzegovina": "\x1b[34m", // blue (alternate)
	// Group C
	"Brazil":   "\x1b[33m", // yellow
	"Morocco":  "\x1b[31m", // red
	"Scotland": "\x1b[34m", // blue
	"Haiti":    "\x1b[36m", // cyan
	// Group D
	"United States": "\x1b[34m", // blue (BBC FullName/ShortName)
	"USA":           "\x1b[34m", // blue (alternate)
	"Paraguay":      "\x1b[31m", // red
	"Australia":     "\x1b[33m", // gold
	"Türkiye":       "\x1b[31m", // red
	"Turkey":        "\x1b[31m", // red (alternate)
	// Group E
	"Germany":     "\x1b[37m", // white
	"Ecuador":     "\x1b[33m", // yellow
	"Curaçao":     "\x1b[34m", // blue
	"Curacao":     "\x1b[34m", // blue (alternate)
	"Ivory Coast": "\x1b[31m", // red
	// Group F
	"Netherlands": "\x1b[31m", // red (orange kit → red)
	"Japan":       "\x1b[31m", // red
	"Sweden":      "\x1b[33m", // yellow
	"Tunisia":     "\x1b[35m", // magenta
	// Group G
	"Belgium":     "\x1b[31m", // red
	"Egypt":       "\x1b[31m", // red
	"Iran":        "\x1b[32m", // green
	"New Zealand": "\x1b[34m", // blue
	// Group H
	"Spain":        "\x1b[31m", // red (La Roja)
	"Uruguay":      "\x1b[36m", // cyan (sky blue)
	"Saudi Arabia": "\x1b[32m", // green
	"Cape Verde":   "\x1b[34m", // blue
	// Group I
	"France":  "\x1b[34m", // blue (Les Bleus)
	"Norway":  "\x1b[31m", // red
	"Senegal": "\x1b[32m", // green
	"Iraq":    "\x1b[31m", // red
	// Group J
	"Argentina": "\x1b[36m", // cyan (sky blue)
	"Algeria":   "\x1b[32m", // green
	"Austria":   "\x1b[31m", // red
	"Jordan":    "\x1b[31m", // red
	// Group K
	"Portugal":   "\x1b[31m", // red
	"Colombia":   "\x1b[33m", // yellow
	"Congo DR":   "\x1b[34m", // blue
	"Uzbekistan": "\x1b[34m", // blue
	// Group L
	"England": "\x1b[37m", // white (Three Lions)
	"Croatia": "\x1b[31m", // red
	"Ghana":   "\x1b[33m", // gold
	"Panama":  "\x1b[31m", // red
}

// sortedKeys returns the keys of m sorted by length descending, so longer
// names are replaced before shorter ones that could be substrings.
func sortedKeys(m map[string]string) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) > len(names[j])
	})
	return names
}

// buildColorMap builds a display-name → ANSI code map from the teams actually
// present in wc. It looks up by ShortName first, then FullName, so BBC API
// abbreviations (e.g. "S Korea", "Bosnia") are resolved automatically.
func buildColorMap(wc *WorldCup) map[string]string {
	m := map[string]string{}
	for _, group := range wc.EventGroups {
		for _, sg := range group.SecondaryGroups {
			for _, ev := range sg.Events {
				for _, side := range []struct{ short, full string }{
					{ev.Home.ShortName, ev.Home.FullName},
					{ev.Away.ShortName, ev.Away.FullName},
				} {
					display := side.short
					if display == "" {
						display = side.full
					}
					if display == "" {
						continue
					}
					if c, ok := teamColors[display]; ok {
						m[display] = c
					} else if c, ok := teamColors[side.full]; ok {
						m[display] = c
					}
				}
			}
		}
	}
	return m
}

// RegisterService registers the wc command handler.
func RegisterService(dg *discordgo.Session, config *Service) {
	svcConfig = config
	dg.AddHandler(invokeCommand)
}

// applyWCColors takes a string formatted as pipe-separated columns (already
// columnized) plus header lines prefixed with "---", wraps it in an ANSI code
// block, and applies team/header colours.
func applyWCColors(formatted string, colorMap map[string]string) string {
	lines := strings.Split(formatted, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "---") {
			lines[i] = "\x1b[1;33m" + line + "\x1b[0m"
		}
	}
	result := strings.Join(lines, "\n")
	for _, name := range sortedKeys(colorMap) {
		result = strings.ReplaceAll(result, name, colorMap[name]+name+"\x1b[0m")
	}
	return fmt.Sprintf("```ansi\n%s```", result)
}

// applyAlertColors wraps watch-mode alert lines in an ANSI code block and
// colours any known team names found anywhere in the text.
func applyAlertColors(formatted string, colorMap map[string]string) string {
	result := formatted
	for _, name := range sortedKeys(colorMap) {
		result = strings.ReplaceAll(result, name, colorMap[name]+name+"\x1b[0m")
	}
	return fmt.Sprintf("```ansi\n%s```", result)
}

// sendColoredMessage sends an ansi-formatted message, splitting into multiple
// messages at line boundaries if the content exceeds Discord's 2000-char limit.
func sendColoredMessage(s *discordgo.Session, channelID, colored string) {
	const maxLen = 1900
	if len(colored) <= maxLen {
		s.ChannelMessageSend(channelID, colored)
		return
	}
	inner := strings.TrimPrefix(colored, "```ansi\n")
	inner = strings.TrimSuffix(inner, "```")
	lines := strings.Split(inner, "\n")
	var chunk []string
	for _, line := range lines {
		candidate := fmt.Sprintf("```ansi\n%s```", strings.Join(append(append([]string{}, chunk...), line), "\n"))
		if len(candidate) > maxLen && len(chunk) > 0 {
			s.ChannelMessageSend(channelID, fmt.Sprintf("```ansi\n%s```", strings.Join(chunk, "\n")))
			chunk = []string{line}
		} else {
			chunk = append(chunk, line)
		}
	}
	if len(chunk) > 0 {
		s.ChannelMessageSend(channelID, fmt.Sprintf("```ansi\n%s```", strings.Join(chunk, "\n")))
	}
}

func buildURL(start, end time.Time) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf(svcConfig.TargetURL, end.Format("2006-01-02"), start.Format("2006-01-02"), today)
}

func fetchData(start, end time.Time) (*WorldCup, error) {
	url := buildURL(start, end)
	var wc WorldCup
	err := web.GetJSON(url, &wc)
	return &wc, err
}

func statusLabel(status, periodLabel, displayTimeUK string) string {
	switch strings.ToLower(status) {
	case "postevent", "result":
		return "FT"
	case "midevent", "live":
		if periodLabel != "" {
			return periodLabel
		}
		return "LIVE"
	default:
		if displayTimeUK != "" {
			return displayTimeUK
		}
		return "TBC"
	}
}

type eventRow struct {
	isHeader bool
	content  string
}

// formatEvents builds the columnized fixture table with group headers
// interleaved. Headers are kept separate from the columnize pass to avoid
// inflating column widths.
func formatEvents(wc *WorldCup) string {
	var rows []eventRow
	var dataOnly []string

	for _, group := range wc.EventGroups {
		headerPrinted := false
		for _, sg := range group.SecondaryGroups {
			for _, ev := range sg.Events {
				if !headerPrinted {
					rows = append(rows, eventRow{isHeader: true, content: group.DisplayLabel})
					headerPrinted = true
				}
				home := ev.Home.ShortName
				if home == "" {
					home = ev.Home.FullName
				}
				away := ev.Away.ShortName
				if away == "" {
					away = ev.Away.FullName
				}
				var score string
				if ev.Home.Score != "" && ev.Away.Score != "" {
					score = fmt.Sprintf("%s - %s", ev.Home.Score, ev.Away.Score)
				} else {
					score = "v"
				}
				date := ev.Date.ShortDate
				if date == "" {
					date = ev.StartDateTime.Format("02/01")
				}
				status := statusLabel(ev.Status, ev.PeriodLabel.Value, ev.Time.DisplayTimeUK)
				data := fmt.Sprintf("%s | %-17s | %-7s | %-17s | %s", date, home, score, away, status)
				rows = append(rows, eventRow{content: data})
				dataOnly = append(dataOnly, data)
			}
		}
	}

	if len(dataOnly) == 0 {
		return ""
	}

	formatted := columnize.SimpleFormat(dataOnly)
	formattedLines := strings.Split(formatted, "\n")

	var out []string
	dataIdx := 0
	for _, r := range rows {
		if r.isHeader {
			out = append(out, fmt.Sprintf("--- %s ---", r.content))
		} else {
			if dataIdx < len(formattedLines) {
				out = append(out, formattedLines[dataIdx])
				dataIdx++
			}
		}
	}
	return strings.Join(out, "\n")
}

func collectActions(wc *WorldCup) map[string]matchSnapshot {
	snap := map[string]matchSnapshot{}
	for _, group := range wc.EventGroups {
		for _, sg := range group.SecondaryGroups {
			for _, ev := range sg.Events {
				home := ev.Home.ShortName
				if home == "" {
					home = ev.Home.FullName
				}
				away := ev.Away.ShortName
				if away == "" {
					away = ev.Away.FullName
				}

				seenKeys := map[string]bool{}
				var actions []actionDetail

				for _, a := range ev.Home.Actions {
					for _, sub := range a.Actions {
						key := fmt.Sprintf("%s:%s:%s", a.PlayerName, sub.TimeLabel.Value, sub.Type)
						seenKeys[key] = true
						actions = append(actions, actionDetail{
							key:        key,
							playerName: a.PlayerName,
							actionType: sub.Type,
							timeLabel:  sub.TimeLabel.Value,
							teamName:   home,
						})
					}
				}

				for _, a := range ev.Away.Actions {
					for _, sub := range a.Actions {
						key := fmt.Sprintf("%s:%s:%s", a.PlayerName, sub.TimeLabel.Value, sub.Type)
						seenKeys[key] = true
						actions = append(actions, actionDetail{
							key:        key,
							playerName: a.PlayerName,
							actionType: sub.Type,
							timeLabel:  sub.TimeLabel.Value,
							teamName:   away,
						})
					}
				}

				snap[ev.Urn] = matchSnapshot{
					homeScore:   ev.Home.Score,
					awayScore:   ev.Away.Score,
					status:      ev.Status,
					periodLabel: ev.PeriodLabel.Value,
					seenKeys:    seenKeys,
					actions:     actions,
				}
			}
		}
	}
	return snap
}

func actionTypeLabel(t string) string {
	switch strings.ToLower(t) {
	case "goal":
		return "GOAL"
	case "own-goal":
		return "OWN GOAL"
	case "penalty-goal":
		return "PENALTY"
	case "yellow-card":
		return "YELLOW CARD"
	case "red-card":
		return "RED CARD"
	default:
		return strings.ToUpper(t)
	}
}

func isGoalType(t string) bool {
	switch strings.ToLower(t) {
	case "goal", "penalty-goal", "own-goal", "own goal", "penalty":
		return true
	}
	return false
}

// buildAllActionLines formats every action in snap as a display line, sorted
// by match time so home and away actions are interleaved chronologically.
// By reading from the current snapshot we always reflect the latest API data,
// including pre-watch actions and any scorer corrections made by the BBC.
func buildAllActionLines(snap matchSnapshot) []string {
	sorted := make([]actionDetail, len(snap.actions))
	copy(sorted, snap.actions)
	sort.Slice(sorted, func(i, j int) bool {
		return parseMatchMinute(sorted[i].timeLabel) < parseMatchMinute(sorted[j].timeLabel)
	})
	var lines []string
	for _, a := range sorted {
		lines = append(lines, fmt.Sprintf("%-12s | %-6s | %s (%s)", actionTypeLabel(a.actionType), a.timeLabel, a.playerName, a.teamName))
	}
	return lines
}

// parseMatchMinute converts a time label like "7'", "90'+2", "90'+6" to an
// integer suitable for sorting. Added time (e.g. +2) is added as a fraction.
func parseMatchMinute(label string) int {
	label = strings.TrimRight(label, "'")
	if idx := strings.Index(label, "'+"); idx >= 0 {
		base := label[:idx]
		extra := label[idx+2:]
		b, _ := strconv.Atoi(base)
		e, _ := strconv.Atoi(extra)
		return b*100 + e
	}
	v, _ := strconv.Atoi(label)
	return v * 100
}

// compareSnapshots returns:
//   - allNewActions: every brand-new action per URN (no score gating)
//   - scoreChanged:  URNs where the score changed this tick
//   - gameEnded:     URNs that transitioned to PostEvent this tick
//   - headlines:     current "Home X - Y Away (status)" string per URN
func compareSnapshots(old, curr map[string]matchSnapshot, wc *WorldCup) (
	allNewActions map[string][]actionDetail,
	scoreChanged map[string]bool,
	gameEnded map[string]bool,
	headlines map[string]string,
) {
	allNewActions = map[string][]actionDetail{}
	scoreChanged = map[string]bool{}
	gameEnded = map[string]bool{}
	headlines = map[string]string{}

	for _, group := range wc.EventGroups {
		for _, sg := range group.SecondaryGroups {
			for _, ev := range sg.Events {
				urn := ev.Urn
				newSnap := curr[urn]
				prevSnap, existed := old[urn]

				home := ev.Home.ShortName
				if home == "" {
					home = ev.Home.FullName
				}
				away := ev.Away.ShortName
				if away == "" {
					away = ev.Away.FullName
				}

				if !existed {
					continue
				}

				if newSnap.homeScore != prevSnap.homeScore || newSnap.awayScore != prevSnap.awayScore {
					scoreChanged[urn] = true
				}

				prevStatus := strings.ToLower(prevSnap.status)
				newStatus := strings.ToLower(newSnap.status)
				if prevStatus != "postevent" && prevStatus != "result" &&
					(newStatus == "postevent" || newStatus == "result") {
					gameEnded[urn] = true
				}

				status := newSnap.periodLabel
				if status == "" {
					status = newSnap.status
				}
				headlines[urn] = fmt.Sprintf("%s  %s - %s  %s  (%s)", home, newSnap.homeScore, newSnap.awayScore, away, status)

				for _, a := range newSnap.actions {
					if !prevSnap.seenKeys[a.key] {
						allNewActions[urn] = append(allNewActions[urn], a)
					}
				}
			}
		}
	}
	return
}

func hasLiveGames(wc *WorldCup) bool {
	for _, group := range wc.EventGroups {
		for _, sg := range group.SecondaryGroups {
			for _, ev := range sg.Events {
				switch strings.ToLower(ev.Status) {
				case "midevent", "live":
					return true
				}
			}
		}
	}
	return false
}

func startWatch(s *discordgo.Session, channelID string) {
	watchMu.Lock()
	if _, exists := watches[channelID]; exists {
		watchMu.Unlock()
		s.ChannelMessageSend(channelID, "```Watch mode is already active.```")
		return
	}
	sess := &watchSession{
		stop:         make(chan struct{}),
		snapshot:     map[string]matchSnapshot{},
		gameMessages: map[string]string{},

		pendingGoalActions: map[string][]actionDetail{},
		pendingGoalKeys:    map[string]map[string]bool{},
	}
	watches[channelID] = sess
	watchMu.Unlock()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Initial fetch to establish baseline
		now := time.Now()
		data, err := fetchData(now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
		if err != nil {
			s.ChannelMessageSend(channelID, fmt.Sprintf("```Watch mode enabled (error on initial fetch: %s)```", err))
		} else {
			watchMu.Lock()
			sess.snapshot = collectActions(data)
			watchMu.Unlock()

			if hasLiveGames(data) {
				formatted := formatEvents(data)
				s.ChannelMessageSend(channelID, "```Watch mode enabled.```")
				sendColoredMessage(s, channelID, applyWCColors(formatted, buildColorMap(data)))
			} else {
				s.ChannelMessageSend(channelID, "```Watch mode enabled. No games currently on.```")
			}
		}

		liveWasActive := false
		_ = liveWasActive
		for {
			select {
			case <-sess.stop:
				return
			case <-ticker.C:
				now := time.Now()
				data, err := fetchData(now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
				if err != nil {
					continue
				}
				newSnap := collectActions(data)
				watchMu.Lock()
				oldSnap := sess.snapshot
				sess.snapshot = newSnap
				watchMu.Unlock()

				allNewActions, scoreChanged, gameEnded, headlines := compareSnapshots(oldSnap, newSnap, data)
				colorMap := buildColorMap(data)

				urnsToPost := map[string]struct{}{}

				watchMu.Lock()
				for urn, actions := range allNewActions {
					for _, a := range actions {
						if isGoalType(a.actionType) {
							if sess.pendingGoalKeys[urn] == nil {
								sess.pendingGoalKeys[urn] = map[string]bool{}
							}
							if !sess.pendingGoalKeys[urn][a.key] {
								sess.pendingGoalActions[urn] = append(sess.pendingGoalActions[urn], a)
								sess.pendingGoalKeys[urn][a.key] = true
							}
						} else {
							urnsToPost[urn] = struct{}{}
						}
					}
				}
				for urn := range scoreChanged {
					if len(sess.pendingGoalActions[urn]) > 0 {
						sess.pendingGoalActions[urn] = nil
						sess.pendingGoalKeys[urn] = nil
						urnsToPost[urn] = struct{}{}
					}
				}
				// Always repost when a game ends (FT)
				for urn := range gameEnded {
					urnsToPost[urn] = struct{}{}
				}
				prevMsgIDs := map[string]string{}
				for urn := range urnsToPost {
					prevMsgIDs[urn] = sess.gameMessages[urn]
				}
				watchMu.Unlock()

				for urn := range urnsToPost {
					// Build output from the full current snapshot so pre-watch actions
					// (e.g. goals before watch started) and BBC corrections are reflected.
					lines := buildAllActionLines(newSnap[urn])
					if len(lines) == 0 {
						continue
					}
					headline := headlines[urn]
					if prevMsgIDs[urn] != "" {
						s.ChannelMessageDelete(channelID, prevMsgIDs[urn])
					}
					body := columnize.SimpleFormat(lines)
					combined := fmt.Sprintf("%s\n\n%s", headline, body)
					msg, err := s.ChannelMessageSend(channelID, applyAlertColors(combined, colorMap))
					if err == nil {
						watchMu.Lock()
						sess.gameMessages[urn] = msg.ID
						watchMu.Unlock()
					}
				}
			}
		}
	}()
}

func stopWatch(s *discordgo.Session, channelID string) {
	watchMu.Lock()
	defer watchMu.Unlock()

	sess, exists := watches[channelID]
	if !exists {
		s.ChannelMessageSend(channelID, "```Watch mode is not active.```")
		return
	}
	close(sess.stop)
	delete(watches, channelID)
	s.ChannelMessageSend(channelID, "```Watch mode disabled.```")
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	parts := strings.Fields(m.Content)
	if len(parts) == 0 || parts[0] != "!wc" {
		return
	}

	switch {
	case len(parts) == 1:
		// Yesterday to tomorrow: scores and fixtures
		now := time.Now()
		data, err := fetchData(now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching WC data: %s", err))
			return
		}
		formatted := formatEvents(data)
		if formatted == "" {
			s.ChannelMessageSend(m.ChannelID, "```No fixtures found.```")
			return
		}
		sendColoredMessage(s, m.ChannelID, applyWCColors(formatted, buildColorMap(data)))

	case len(parts) >= 2 && parts[1] == "fixtures":
		// One week of upcoming fixtures
		now := time.Now()
		data, err := fetchData(now, now.AddDate(0, 0, 7))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching WC fixtures: %s", err))
			return
		}
		formatted := formatEvents(data)
		if formatted == "" {
			s.ChannelMessageSend(m.ChannelID, "```No fixtures found.```")
			return
		}
		sendColoredMessage(s, m.ChannelID, applyWCColors(formatted, buildColorMap(data)))

	case len(parts) >= 3 && parts[1] == "watch" && parts[2] == "on":
		startWatch(s, m.ChannelID)

	case len(parts) >= 3 && parts[1] == "watch" && parts[2] == "off":
		stopWatch(s, m.ChannelID)
	}
}
