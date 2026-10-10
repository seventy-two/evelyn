package autotrader

import (
	"fmt"
	"net/url"
	"sort"

	log "github.com/sirupsen/logrus"
)

// filterNames maps search page URL parameters to the API's filter names.
var filterNames = map[string]string{
	"make":         "make",
	"model":        "model",
	"postcode":     "postcode",
	"transmission": "transmission",
	"minPrice":     "min_price",
	"price-from":   "min_price",
	"maxPrice":     "max_price",
	"price-to":     "max_price",
	"year-from":    "min_year_manufactured",
	"year-to":      "max_year_manufactured",
}

// ignoredParams are search page URL parameters that aren't filters.
var ignoredParams = map[string]bool{
	"channel": true,
	"flrfc":   true,
	"sort":    true,
	"page":    true,
}

type filter struct {
	Filter   string   `json:"filter"`
	Selected []string `json:"selected"`
}

type search struct {
	channel string
	make    string
	model   string
	filters []filter
}

// parseSearch turns an AutoTrader search page URL into API filters.
func parseSearch(raw string) (*search, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing search URL: %w", err)
	}
	q := u.Query()

	s := &search{
		channel: q.Get("channel"),
		make:    q.Get("make"),
		model:   q.Get("model"),
		// The API requires this; "total" means prices aren't monthly.
		filters: []filter{{Filter: "price_search_type", Selected: []string{"total"}}},
	}
	if s.channel == "" {
		s.channel = "cars"
	}

	params := make([]string, 0, len(q))
	for p := range q {
		params = append(params, p)
	}
	sort.Strings(params)
	for _, p := range params {
		name, ok := filterNames[p]
		if !ok {
			if !ignoredParams[p] {
				log.Warnf("autotrader: ignoring unsupported search parameter %q", p)
			}
			continue
		}
		s.filters = append(s.filters, filter{Filter: name, Selected: q[p]})
	}
	if s.make == "" {
		return nil, fmt.Errorf("search URL has no make")
	}
	return s, nil
}

// name describes what's being searched for, e.g. "Mazda3".
func (s *search) name() string {
	if s.model != "" {
		return s.model
	}
	return s.make
}
