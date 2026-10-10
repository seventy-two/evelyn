// Package autotrader watches an AutoTrader search and posts newly listed
// cars to Discord.
package autotrader

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	gatewayURL = "https://www.autotrader.co.uk/at-gateway?opname=SearchResultsListingsGridQuery"
	siteURL    = "https://www.autotrader.co.uk"
	userAgent  = "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0"
	seenKey    = "autotrader:seen"
)

// listingsQuery asks for only the fields an alert needs.
const listingsQuery = `query SearchResultsListingsGridQuery($filters: [FilterInput!]!, $channel: Channel!, $page: Int, $sortBy: SearchResultsSort, $listingType: [ListingType!], $searchId: String!, $featureFlags: [FeatureFlag]) {
  searchResults(input: {facets: [], filters: $filters, channel: $channel, page: $page, sortBy: $sortBy, listingType: $listingType, searchId: $searchId, featureFlags: $featureFlags}) {
    listings { ... on SearchListing { type advertId title subTitle attentionGrabber price vehicleLocation sellerType images badges { type displayText } } }
  }
}`

// Service is the configuration for the AutoTrader watcher. It does nothing
// unless SearchURL is set.
type Service struct {
	SearchURL string
	ChannelID string
	UserID    string
	Interval  time.Duration
	Store     Store
}

// Store keeps which adverts have already been seen.
type Store interface {
	GetValue(key string, v interface{}) (bool, error)
	SetValue(key string, v interface{}) error
}

// RegisterService starts checking the search in the background.
func RegisterService(dg *discordgo.Session, config *Service) {
	if config.SearchURL == "" {
		return
	}
	s, err := parseSearch(config.SearchURL)
	if err != nil {
		log.Errorf("autotrader: %s", err)
		return
	}
	w := &watcher{
		search:  s,
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		gateway: gatewayURL,
		send: func(content string, embed *discordgo.MessageEmbed) error {
			_, err := dg.ChannelMessageSendComplex(config.ChannelID, &discordgo.MessageSend{
				Content:         content,
				Embeds:          []*discordgo.MessageEmbed{embed},
				AllowedMentions: &discordgo.MessageAllowedMentions{Users: []string{config.UserID}},
			})
			return err
		},
	}
	go w.run()
}

type watcher struct {
	search  *search
	config  *Service
	client  *http.Client
	gateway string
	send    func(content string, embed *discordgo.MessageEmbed) error
}

func (w *watcher) run() {
	for {
		if err := w.check(); err != nil {
			// If AutoTrader starts blocking us, just try again next time.
			log.Errorf("autotrader: %s", err)
		}
		time.Sleep(w.config.Interval)
	}
}

// check fetches the search and posts any cars not seen before. The first
// check only records what's already listed.
func (w *watcher) check() error {
	listings, err := w.fetch()
	if err != nil {
		return err
	}

	var seen []string
	known, err := w.config.Store.GetValue(seenKey, &seen)
	if err != nil {
		return fmt.Errorf("loading seen adverts: %w", err)
	}
	isSeen := map[string]bool{}
	for _, id := range seen {
		isSeen[id] = true
	}

	for _, l := range listings {
		if isSeen[l.AdvertID] {
			continue
		}
		if known {
			content := fmt.Sprintf("<@%s> new %s on AutoTrader", w.config.UserID, w.search.name())
			if err := w.send(content, l.embed()); err != nil {
				// Leave it unseen so the next check tries again.
				log.Errorf("autotrader: posting advert %s: %s", l.AdvertID, err)
				continue
			}
		}
		isSeen[l.AdvertID] = true
		seen = append(seen, l.AdvertID)
	}
	if !known {
		log.Infof("autotrader: watching %s, %d already listed", w.search.name(), len(listings))
	}
	return w.config.Store.SetValue(seenKey, seen)
}

type listing struct {
	Type             string   `json:"type"`
	AdvertID         string   `json:"advertId"`
	Title            string   `json:"title"`
	SubTitle         string   `json:"subTitle"`
	AttentionGrabber string   `json:"attentionGrabber"`
	Price            string   `json:"price"`
	VehicleLocation  string   `json:"vehicleLocation"`
	SellerType       string   `json:"sellerType"`
	Images           []string `json:"images"`
	Badges           []struct {
		Type        string `json:"type"`
		DisplayText string `json:"displayText"`
	} `json:"badges"`
}

type graphQLResponse struct {
	Data struct {
		SearchResults struct {
			Listings []listing `json:"listings"`
		} `json:"searchResults"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// fetch returns the first page of the search's results, newest first. Only
// genuine results count: promoted adverts and suggestions don't necessarily
// match the search.
func (w *watcher) fetch() ([]listing, error) {
	body, err := json.Marshal([]map[string]interface{}{{
		"operationName": "SearchResultsListingsGridQuery",
		"query":         listingsQuery,
		"variables": map[string]interface{}{
			"filters":      w.search.filters,
			"channel":      w.search.channel,
			"page":         1,
			"sortBy":       "most_recent",
			"listingType":  nil,
			"searchId":     newSearchID(),
			"featureFlags": []string{},
		},
	}})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, w.gateway, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", siteURL)
	req.Header.Set("Referer", siteURL+"/car-search")
	req.Header.Set("x-sauron-app-name", "sauron-search-results-app")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned %s", resp.Status)
	}

	var results []graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decoding search results: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("empty search response")
	}
	if len(results[0].Errors) > 0 {
		return nil, fmt.Errorf("search failed: %s", results[0].Errors[0].Message)
	}

	var listings []listing
	for _, l := range results[0].Data.SearchResults.Listings {
		if l.Type == "NATURAL_LISTING" && l.AdvertID != "" {
			listings = append(listings, l)
		}
	}
	return listings, nil
}

// embed describes a listing as a Discord card.
func (l *listing) embed() *discordgo.MessageEmbed {
	title := strings.TrimSpace(l.Title + " · " + l.Price)
	e := &discordgo.MessageEmbed{
		Title:       title,
		URL:         siteURL + "/car-details/" + l.AdvertID,
		Description: l.SubTitle,
		Color:       0x00A1E4,
	}
	if l.AttentionGrabber != "" {
		e.Description += "\n*" + strings.TrimSpace(l.AttentionGrabber) + "*"
	}

	field := func(name, value string) {
		if value != "" {
			e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: name, Value: value, Inline: true})
		}
	}
	var notes []string
	for _, b := range l.Badges {
		switch {
		case b.Type == "REGISTERED_YEAR":
			field("Year", b.DisplayText)
		case b.Type == "MILEAGE":
			field("Mileage", b.DisplayText)
		case b.Type == "WRITE_OFF_CATEGORY":
			notes = append(notes, "⚠️ "+b.DisplayText+" write-off")
		case strings.HasPrefix(b.Type, "PI_"):
			notes = append(notes, b.DisplayText)
		}
	}
	field("Location", l.VehicleLocation)
	field("Seller", sellerName(l.SellerType))
	if len(notes) > 0 {
		field("Notes", strings.Join(notes, "\n"))
	}
	if len(l.Images) > 0 {
		e.Image = &discordgo.MessageEmbedImage{URL: strings.Replace(l.Images[0], "{resize}", "w800h600", 1)}
	}
	return e
}

func sellerName(t string) string {
	switch t {
	case "PRIVATE":
		return "Private"
	case "TRADE":
		return "Dealer"
	default:
		return ""
	}
}

// newSearchID makes a random UUID, as the site sends with each search.
func newSearchID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
