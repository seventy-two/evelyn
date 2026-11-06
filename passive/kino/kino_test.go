package kino

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore map[string][]byte

func (f fakeStore) GetValue(key string, v interface{}) (bool, error) {
	b, ok := f[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(b, v)
}

func (f fakeStore) SetValue(key string, v interface{}) error {
	b, err := json.Marshal(v)
	f[key] = b
	return err
}

// fakeKino serves a feed of logs and a member list.
type fakeKino struct {
	logs    []logEntry
	members []user
}

func (k *fakeKino) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/users":
			json.NewEncoder(w).Encode(map[string]any{"members": k.members})
		case "/api/feed":
			var since int64
			poll := r.URL.Query().Has("since")
			if poll {
				json.Unmarshal([]byte(r.URL.Query().Get("since")), &since)
			}
			latest := int64(0)
			if len(k.logs) > 0 {
				latest = k.logs[len(k.logs)-1].ID
			}
			items := []logEntry{}
			if poll {
				for _, l := range k.logs {
					if l.ID > since {
						items = append(items, l)
					}
				}
				latest = since
				if len(items) > 0 {
					latest = items[len(items)-1].ID
				}
			}
			json.NewEncoder(w).Encode(map[string]any{"items": items, "latest": latest})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	})
}

func heat(id int64, who string) logEntry {
	return logEntry{ID: id, URL: "https://k/log/1", User: user{Username: strings.ToLower(who), Name: who, URL: "https://k/u/" + who},
		Film: film{Label: "Heat (1995)", Poster: "https://img/heat.jpg"}, Verb: "reviewed", Stars: "★★★★½", Liked: true,
		WatchedText: "6 Nov 2026", Review: "The diner scene.", Spoilers: true}
}

func newWatcher(t *testing.T, k *fakeKino) (*watcher, *[]*discordgo.MessageSend, fakeStore) {
	srv := httptest.NewServer(k.handler(t))
	t.Cleanup(srv.Close)
	store := fakeStore{}
	var sent []*discordgo.MessageSend
	w := &watcher{config: &Service{URL: srv.URL, Store: store}, client: srv.Client(),
		send: func(m *discordgo.MessageSend) error { sent = append(sent, m); return nil }}
	return w, &sent, store
}

func TestFeedPostsOnlyNewLogs(t *testing.T) {
	k := &fakeKino{logs: []logEntry{heat(1, "Mo")}}
	w, sent, _ := newWatcher(t, k)
	require.NoError(t, w.checkFeed())
	assert.Empty(t, *sent, "the first check only records the cursor")

	k.logs = append(k.logs, heat(2, "Jo"))
	require.NoError(t, w.checkFeed())
	require.Len(t, *sent, 1)
	e := (*sent)[0].Embeds[0]
	assert.Equal(t, "Jo reviewed", e.Author.Name)
	assert.Equal(t, "Heat (1995)", e.Title)
	assert.Equal(t, "★★★★½  ·  ♥  ·  watched 6 Nov 2026\n\n||The diner scene.||", e.Description)
	assert.Equal(t, "https://img/heat.jpg", e.Thumbnail.URL)

	require.NoError(t, w.checkFeed())
	assert.Len(t, *sent, 1, "nothing new, nothing posted")
}

func TestFeedStartsFromAnEmptyFeed(t *testing.T) {
	k := &fakeKino{}
	w, sent, store := newWatcher(t, k)
	require.NoError(t, w.checkFeed())
	assert.Equal(t, "0", string(store[cursorKey]))
	require.NoError(t, w.checkFeed())
	assert.Empty(t, *sent)
	k.logs = append(k.logs, heat(1, "Mo"))
	require.NoError(t, w.checkFeed())
	require.Len(t, *sent, 1, "the first log ever is posted")
}

func TestFeedCollapsesABatch(t *testing.T) {
	k := &fakeKino{logs: []logEntry{heat(1, "Mo")}}
	w, sent, _ := newWatcher(t, k)
	require.NoError(t, w.checkFeed())
	for i := int64(2); i <= 6; i++ {
		k.logs = append(k.logs, heat(i, "Mo"))
	}
	k.logs = append(k.logs, heat(7, "Jo"))
	require.NoError(t, w.checkFeed())
	require.Len(t, *sent, 2)
	assert.Equal(t, "Mo logged 5 films", (*sent)[0].Embeds[0].Author.Name)
	assert.Equal(t, "Jo reviewed", (*sent)[1].Embeds[0].Author.Name)
}

func TestMembers(t *testing.T) {
	k := &fakeKino{members: []user{{Username: "megz", Name: "megz", URL: "https://k/u/megz"}}}
	w, sent, _ := newWatcher(t, k)
	require.NoError(t, w.checkMembers())
	assert.Empty(t, *sent, "existing members aren't announced")

	k.members = append(k.members, user{Username: "jo_b", Name: "Jo*", URL: "https://k/u/jo_b"})
	require.NoError(t, w.checkMembers())
	require.Len(t, *sent, 1)
	assert.Equal(t, `🎟️ **Jo\*** (jo\_b) just joined kino: <https://k/u/jo_b>`, (*sent)[0].Content)

	require.NoError(t, w.checkMembers())
	assert.Len(t, *sent, 1)
}

func TestValidURL(t *testing.T) {
	assert.True(t, validURL("http://127.0.0.1:6100"))
	assert.False(t, validURL("127.0.0.1:6100"))
}
