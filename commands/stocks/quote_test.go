package stocks

import (
	"encoding/json"
	"testing"
)

// cnbcAnthropic is a trimmed real response from CNBC's symbol search for
// "anthropic" (23 Nov 2026): metadata first, then the ETF ranked above ANT.
const cnbcAnthropic = `[{"TotalMatchAvailable":"10","TotalMatchReturned":"10"},
{"issueType":"ETF","countryCode":"US","exchangeName":"NYSE Arca","symbolName":"ANTW","companyName":"Anthropic AI Lab Ecosystem ETF"},
{"issueType":"STOCK","countryCode":"US","exchangeName":"NASDAQ","symbolName":"ANT","companyName":"Anthropic, PBC"},
{"issueType":"INDEX","countryCode":"US","exchangeName":"Exchange","symbolName":"ANTHR.FG","companyName":"Anthropic"}]`

func TestPickSymbol(t *testing.T) {
	var anthropic []LookupMatch
	if err := json.Unmarshal([]byte(cnbcAnthropic), &anthropic); err != nil {
		t.Fatal(err)
	}
	etfs := []LookupMatch{
		{Symbol: "ANTC", IssueType: "ETF", CountryCode: "US"},
		{Symbol: "ANTW", IssueType: "ETF", CountryCode: "US"},
	}
	foreignOnly := []LookupMatch{{Symbol: "TSCO-GB", IssueType: "STOCK", CountryCode: "GB"}}

	tests := []struct {
		name    string
		query   string
		matches []LookupMatch
		want    string
	}{
		{"name prefers US stock over ETF", "anthropic", anthropic, "ANT"},
		{"exact ticker wins, any case", "antw", anthropic, "ANTW"},
		{"multi-word query", "anthropic+pbc", anthropic, "ANT"},
		{"no US stock, take US listing", "ant+etf", etfs, "ANTC"},
		{"only foreign listing", "tesco", foreignOnly, "TSCO-GB"},
		{"metadata only", "zzz", anthropic[:1], ""},
		{"no matches", "zzz", nil, ""},
	}
	for _, tt := range tests {
		if got := pickSymbol(tt.query, tt.matches); got != tt.want {
			t.Errorf("%s: pickSymbol(%q) = %q, want %q", tt.name, tt.query, got, tt.want)
		}
	}
}

func TestSearchVariants(t *testing.T) {
	if got := searchVariants("tesla"); len(got) != 1 || got[0] != "tesla" {
		t.Errorf("single word: got %q", got)
	}
	got := searchVariants("rolls+royce")
	want := []string{"rolls+royce", "rolls-royce", "rolls"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("variant %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
