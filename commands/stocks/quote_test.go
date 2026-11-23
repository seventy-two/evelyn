package stocks

import "testing"

func TestPickSymbol(t *testing.T) {
	anthropic := []LookupMatch{
		{Symbol: "ANTH.TRT", Region: "Toronto"},
		{Symbol: "ANT", Region: "United States"},
	}
	tickerFirst := []LookupMatch{
		{Symbol: "ANTA", Region: "United States"},
		{Symbol: "ANT", Region: "United States"},
	}
	foreignOnly := []LookupMatch{{Symbol: "TSCO.LON", Region: "United Kingdom"}}

	tests := []struct {
		name    string
		query   string
		matches []LookupMatch
		want    string
	}{
		{"name prefers US listing", "anthropic", anthropic, "ANT"},
		{"exact ticker wins", "ant", tickerFirst, "ANT"},
		{"only foreign listing", "tesco", foreignOnly, "TSCO.LON"},
		{"multi-word query", "anthropic+pbc", anthropic, "ANT"},
		{"no matches", "zzz", nil, ""},
	}
	for _, tt := range tests {
		if got := pickSymbol(tt.query, tt.matches); got != tt.want {
			t.Errorf("%s: pickSymbol(%q) = %q, want %q", tt.name, tt.query, got, tt.want)
		}
	}
}
