package hearth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memStore map[string]any

func (m memStore) GetValue(k string, v interface{}) (bool, error) {
	s, ok := m[k]
	if ok {
		switch p := v.(type) {
		case *string:
			*p = s.(string)
		case *bool:
			*p = s.(bool)
		}
	}
	return ok, nil
}

func (m memStore) SetValue(k string, v interface{}) error {
	m[k] = v
	return nil
}

func setup(t *testing.T, live bool) (*poster, *[]*discordgo.MessageSend, memStore) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/light/status":
			if live {
				w.Write([]byte(`{"count":1234,"target":2000,"live":true,"start":"2026-11-24T21:45:00Z"}`))
			} else {
				w.Write([]byte(`{"count":0,"target":2000,"live":false}`))
			}
		case "/snapshot.png":
			w.Write([]byte("png"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	london, _ := time.LoadLocation("Europe/London")
	store := memStore{launchedKey: true}
	var sent []*discordgo.MessageSend
	p := &poster{
		config: &Service{URL: srv.URL, Hour: 10, Store: store},
		client: srv.Client(), london: london,
		send: func(m *discordgo.MessageSend) error { sent = append(sent, m); return nil },
	}
	return p, &sent, store
}

func TestPostsOncePerDayAfterTen(t *testing.T) {
	p, sent, _ := setup(t, true)
	at := func(s string) time.Time { tm, _ := time.ParseInLocation("2006-01-02 15:04", s, p.london); return tm }
	require.NoError(t, p.tick(at("2026-12-01 09:59")))
	assert.Empty(t, *sent)
	require.NoError(t, p.tick(at("2026-12-01 10:00")))
	require.Len(t, *sent, 1)
	assert.Contains(t, (*sent)[0].Content, "day 7")
	assert.Contains(t, (*sent)[0].Content, "1,234 messages")
	assert.Contains(t, (*sent)[0].Content, "61% lit")
	assert.Equal(t, "hearth.png", (*sent)[0].Files[0].Name)
	require.NoError(t, p.tick(at("2026-12-01 10:05")))
	assert.Len(t, *sent, 1, "only once a day")
	require.NoError(t, p.tick(at("2026-12-02 10:01")))
	assert.Len(t, *sent, 2)
}

func TestNoLateCatchUp(t *testing.T) {
	p, sent, store := setup(t, true)
	tm, _ := time.ParseInLocation("2006-01-02 15:04", "2026-12-01 19:00", p.london)
	require.NoError(t, p.tick(tm))
	assert.Empty(t, *sent)
	assert.Equal(t, "2026-12-01", store[postedKey])
}

func TestQuietUntilLive(t *testing.T) {
	p, sent, _ := setup(t, false)
	require.NoError(t, p.post(time.Now()))
	assert.Empty(t, *sent)
}

func TestDarkCaption(t *testing.T) {
	assert.Contains(t, Caption(Status{Count: 0, Target: 2000, Live: true}, time.Now()), "the living room is dark")
}

func TestLaunchPostOnceWhenLive(t *testing.T) {
	p, sent, store := setup(t, true)
	delete(store, launchedKey)
	tm, _ := time.ParseInLocation("2006-01-02 15:04", "2026-11-24 21:50", p.london)
	require.NoError(t, p.tick(tm))
	require.Len(t, *sent, 1)
	require.NoError(t, p.tick(tm.Add(time.Minute)))
	assert.Len(t, *sent, 1, "launch post only once")
	require.NoError(t, p.tick(tm.Add(13*time.Hour))) // next day, 10:50
	assert.Len(t, *sent, 2)
}

func TestNoLaunchPostBeforeLive(t *testing.T) {
	p, sent, store := setup(t, false)
	delete(store, launchedKey)
	tm, _ := time.ParseInLocation("2006-01-02 15:04", "2026-11-24 21:50", p.london)
	require.NoError(t, p.tick(tm))
	assert.Empty(t, *sent)
	_, launched := store[launchedKey]
	assert.False(t, launched)
}
