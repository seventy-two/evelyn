package dictionary

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWOTD = `{"word":"stewartry",
"definitions":[
	{"source":"century","text":"In Scotland, a jurisdiction over a certain extent of territory.","partOfSpeech":"noun"},
	{"source":"gcide","text":"An overseer or superintendent.","partOfSpeech":"noun"},
	{"source":"gcide","text":"The office of a steward; stewardship.","partOfSpeech":"noun"},
	{"source":"gcide","text":"In Scotland, the jurisdiction of a steward.","partOfSpeech":"noun"}],
"examples":[{"title":"Redgauntlet","text":"neither county nor stewartry, as we say"}],
"note":"The word 'stewartry' is related to 'steward'."}`

func TestWOTDEmbed(t *testing.T) {
	data := &wotdResponse{}
	require.NoError(t, json.Unmarshal([]byte(testWOTD), data))

	e := wotdEmbed(data)
	assert.Equal(t, "stewartry", e.Title)
	assert.Equal(t, "https://www.wordnik.com/words/stewartry", e.URL)
	assert.Equal(t, "1. *noun* In Scotland, a jurisdiction over a certain extent of territory.\n"+
		"2. *noun* An overseer or superintendent.\n"+
		"3. *noun* The office of a steward; stewardship.", e.Description)
	require.Len(t, e.Fields, 2)
	assert.Equal(t, "> neither county nor stewartry, as we say\n— *Redgauntlet*", e.Fields[0].Value)
	assert.Equal(t, "The word 'stewartry' is related to 'steward'.", e.Fields[1].Value)
}

func TestWOTDEmbedMinimal(t *testing.T) {
	e := wotdEmbed(&wotdResponse{Word: "word"})
	assert.Equal(t, "word", e.Title)
	assert.Empty(t, e.Description)
	assert.Empty(t, e.Fields)
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", truncate("abc", 3))
	assert.Equal(t, "ab…", truncate("abcd", 3))
}
