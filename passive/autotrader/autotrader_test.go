package autotrader

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSearchURL = "https://www.autotrader.co.uk/car-search?channel=cars&flrfc=1&make=Mazda&maxPrice=14000&model=Mazda3&postcode=SW1A%201AA&sort=relevance&transmission=Automatic&year-from=2021"

func TestParseSearch(t *testing.T) {
	s, err := parseSearch(testSearchURL)
	require.NoError(t, err)
	assert.Equal(t, "cars", s.channel)
	assert.Equal(t, "Mazda3", s.name())
	assert.Equal(t, []filter{
		{Filter: "price_search_type", Selected: []string{"total"}},
		{Filter: "make", Selected: []string{"Mazda"}},
		{Filter: "max_price", Selected: []string{"14000"}},
		{Filter: "model", Selected: []string{"Mazda3"}},
		{Filter: "postcode", Selected: []string{"SW1A 1AA"}},
		{Filter: "transmission", Selected: []string{"Automatic"}},
		{Filter: "min_year_manufactured", Selected: []string{"2021"}},
	}, s.filters)

	_, err = parseSearch("https://www.autotrader.co.uk/car-search?postcode=SW1A")
	assert.Error(t, err, "a search needs a make")
}

// fakeStore is an in-memory Store.
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

func listingJSON(id, typ string) string {
	return `{"type":"` + typ + `","advertId":"` + id + `","title":"Mazda Mazda3","subTitle":"2.0 Sport Lux Auto","price":"£13,900","vehicleLocation":"Aldershot (82 miles)","sellerType":"PRIVATE","attentionGrabber":"Full service history","images":["https://m.atcdn.co.uk/a/media/{resize}/abc.jpg"],"badges":[{"type":"MILEAGE","displayText":"36,500 miles"},{"type":"REGISTERED_YEAR","displayText":"2022 (22 reg)"},{"type":"WRITE_OFF_CATEGORY","displayText":"Cat N"},{"type":"PI_GOOD","displayText":"Good price"}]}`
}

// gateway serves search results with the given listings and records requests.
func gateway(t *testing.T, listings *[]string, requests *[]map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		*requests = append(*requests, body[0])
		io.WriteString(w, `[{"data":{"searchResults":{"listings":[`+strings.Join(*listings, ",")+`]}}}]`)
	}))
}

func newTestWatcher(t *testing.T, gw string, store Store, sent *[]string, failSend *bool) *watcher {
	s, err := parseSearch(testSearchURL)
	require.NoError(t, err)
	return &watcher{
		search:  s,
		config:  &Service{UserID: "42", Store: store},
		client:  http.DefaultClient,
		gateway: gw,
		send: func(content string, e *discordgo.MessageEmbed) error {
			if *failSend {
				return errors.New("discord down")
			}
			*sent = append(*sent, content+" | "+e.URL)
			return nil
		},
	}
}

func TestCheckAlertsOnlyNewListings(t *testing.T) {
	listings := []string{listingJSON("1", "NATURAL_LISTING"), listingJSON("2", "NATURAL_LISTING"), listingJSON("ad", "PROMOTED_LISTING")}
	var requests []map[string]interface{}
	srv := gateway(t, &listings, &requests)
	defer srv.Close()

	var sent []string
	failSend := false
	store := fakeStore{}
	w := newTestWatcher(t, srv.URL, store, &sent, &failSend)

	// First check: record what's there without alerting.
	require.NoError(t, w.check())
	assert.Empty(t, sent)

	// A new car is listed.
	listings = append([]string{listingJSON("3", "NATURAL_LISTING")}, listings...)
	require.NoError(t, w.check())
	assert.Equal(t, []string{"<@42> new Mazda3 on AutoTrader | https://www.autotrader.co.uk/car-details/3"}, sent)

	// Nothing new: no alert. Promoted adverts never alert.
	require.NoError(t, w.check())
	assert.Len(t, sent, 1)

	// If posting fails, the car is retried next time.
	listings = append([]string{listingJSON("4", "NATURAL_LISTING")}, listings...)
	failSend = true
	require.NoError(t, w.check())
	failSend = false
	require.NoError(t, w.check())
	assert.Equal(t, "<@42> new Mazda3 on AutoTrader | https://www.autotrader.co.uk/car-details/4", sent[1])
	assert.Len(t, sent, 2)

	vars := requests[0]["variables"].(map[string]interface{})
	assert.Equal(t, "most_recent", vars["sortBy"])
	assert.Equal(t, "cars", vars["channel"])
	assert.Len(t, vars["filters"], 7)
}

func TestCheckErrors(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		"blocked": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "challenge", http.StatusForbidden)
		},
		"graphql error": func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, `[{"errors":[{"message":"Some input filters were invalid"}]}]`)
		},
		"not json": func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, `<html>Just a moment...</html>`)
		},
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(handler)
			defer srv.Close()
			var sent []string
			failSend := false
			store := fakeStore{}
			w := newTestWatcher(t, srv.URL, store, &sent, &failSend)
			assert.Error(t, w.check())
			assert.Empty(t, store, "a failed check doesn't record anything")
		})
	}
}

func TestEmbed(t *testing.T) {
	var l listing
	require.NoError(t, json.Unmarshal([]byte(listingJSON("202610046662866", "NATURAL_LISTING")), &l))
	e := l.embed()
	assert.Equal(t, "Mazda Mazda3 · £13,900", e.Title)
	assert.Equal(t, "https://www.autotrader.co.uk/car-details/202610046662866", e.URL)
	assert.Equal(t, "2.0 Sport Lux Auto\n*Full service history*", e.Description)
	assert.Equal(t, "https://m.atcdn.co.uk/a/media/w800h600/abc.jpg", e.Image.URL)

	fields := map[string]string{}
	for _, f := range e.Fields {
		fields[f.Name] = f.Value
	}
	assert.Equal(t, map[string]string{
		"Mileage":  "36,500 miles",
		"Year":     "2022 (22 reg)",
		"Location": "Aldershot (82 miles)",
		"Seller":   "Private",
		"Notes":    "⚠️ Cat N write-off\nGood price",
	}, fields)
}

func TestNewSearchID(t *testing.T) {
	id := newSearchID()
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, id)
	assert.NotEqual(t, id, newSearchID())
}
