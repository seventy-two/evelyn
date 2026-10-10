package nfl

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testScoreboard = `{"events":[
{"date":"2026-10-09T00:15Z","competitions":[{
	"competitors":[
		{"homeAway":"home","score":"16","team":{"id":"6","shortDisplayName":"Cowboys"}},
		{"homeAway":"away","score":"24","team":{"id":"27","shortDisplayName":"Buccaneers"}}],
	"status":{"type":{"state":"post","shortDetail":"Final"}},
	"broadcasts":[{"names":["Prime Video"]}]}]},
{"date":"2026-10-11T17:00Z","competitions":[{
	"competitors":[
		{"homeAway":"home","score":"10","team":{"id":"9","shortDisplayName":"Packers"}},
		{"homeAway":"away","score":"7","team":{"id":"3","shortDisplayName":"Bears"}}],
	"status":{"type":{"state":"in","shortDetail":"8:24 - 2nd"}},
	"situation":{"downDistanceText":"3rd & 4 at GB 35","possession":"3"},
	"broadcasts":[{"names":["Fox"]}]}]},
{"date":"2026-10-11T13:30Z","competitions":[{
	"competitors":[
		{"homeAway":"home","score":"0","team":{"id":"30","shortDisplayName":"Jaguars"}},
		{"homeAway":"away","score":"0","team":{"id":"21","shortDisplayName":"Eagles"}}],
	"status":{"type":{"state":"pre","shortDetail":"10/11 - 9:30 AM EDT"}},
	"broadcasts":[{"names":["NFL Net"]}]}]}
]}`

func TestGameLines(t *testing.T) {
	sb := &scoreboard{}
	require.NoError(t, json.Unmarshal([]byte(testScoreboard), sb))

	assert.Equal(t, []string{
		"24 | Buccaneers | @ | Cowboys | 16 | Final",
		"7 | >Bears | @ | Packers | 10 | 8:24 - 2nd 3rd & 4 at GB 35",
		" | Eagles | @ | Jaguars |  | Sun 14:30 NFL Net",
	}, gameLines(sb))
}
