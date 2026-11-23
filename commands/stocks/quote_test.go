package stocks

import (
	"encoding/json"
	"strings"
	"testing"
)

// cnbcAnthropic is a trimmed real response from CNBC's symbol search for
// "anthropic" (23 Nov 2026): metadata first, then the ETF ranked above ANT.
const cnbcAnthropic = `[{"TotalMatchAvailable":"10","TotalMatchReturned":"10"},
{"issueType":"ETF","countryCode":"US","exchangeName":"NYSE Arca","symbolName":"ANTW","companyName":"Anthropic AI Lab Ecosystem ETF"},
{"issueType":"STOCK","countryCode":"US","exchangeName":"NASDAQ","symbolName":"ANT","companyName":"Anthropic, PBC"},
{"issueType":"INDEX","countryCode":"US","exchangeName":"Exchange","symbolName":"ANTHR.FG","companyName":"Anthropic"}]`

func m(issuer, typ, country, symbol string) LookupMatch {
	return LookupMatch{IssuerID: issuer, IssueType: typ, CountryCode: country, Symbol: symbol}
}

func TestPickSymbol(t *testing.T) {
	var anthropic []LookupMatch
	if err := json.Unmarshal([]byte(cnbcAnthropic), &anthropic); err != nil {
		t.Fatal(err)
	}
	// Trimmed real CNBC results, 23 Nov 2026, in CNBC's order.
	tesco := []LookupMatch{
		m("34901", "STOCK", "GB", "TSCO-GB"), m("34901", "STOCK", "DE", "TCO0-FF"),
		m("34901", "STOCK", "US", "TSCDF"), m("34901", "STOCK", "US", "TSCDY"),
	}
	apple := []LookupMatch{
		m("", "STOCK", "CN", "3020-SZ"), m("8016", "STOCK", "US", "AAPL"),
		m("155385", "CSRT", "US", "APLE"), m("8016", "STOCK", "GB", "0R2V-GB"),
	}
	bp := []LookupMatch{
		m("11768", "STOCK", "US", "BP"), m("11768", "STOCK", "GB", "BP.-GB"),
		m("217103", "STOCK", "US", "BPAC"),
	}
	toyota := []LookupMatch{
		m("6371", "STOCK", "US", "TM"), m("6371", "STOCK", "JP", "7203.T-JP"),
	}
	nestle := []LookupMatch{
		m("26357", "STOCK", "CH", "NES.N-CH"), m("26357", "STOCK", "US", "NSRGY"),
	}
	ukOnly := []LookupMatch{
		m("1", "STOCK", "CN", "000001-SZ"), m("2", "STOCK", "GB", "ABC-GB"),
		m("2", "STOCK", "DE", "ABC-FF"),
	}
	etfs := []LookupMatch{m("9", "ETF", "GB", "3LRR-GB"), m("8", "ETF", "US", "ANTC")}
	astra := []LookupMatch{
		{IssuerID: "18664", IssueType: "STOCK", CountryCode: "US", Symbol: "GILD", Name: "Gilead Sciences, Inc."},
		{IssuerID: "10594", IssueType: "STOCK", CountryCode: "GB", Symbol: "AZN-GB", Name: "AstraZeneca PLC"},
		{IssuerID: "10594", IssueType: "STOCK", CountryCode: "US", Symbol: "AZN", Name: "AstraZeneca PLC"},
	}
	novo := []LookupMatch{
		{IssuerID: "157769", IssueType: "STOCK", CountryCode: "US", Symbol: "NVX", Name: "NOVONIX Ltd"},
		{IssuerID: "26780", IssueType: "STOCK", CountryCode: "US", Symbol: "NVO", Name: "Novo Nordisk A/S"},
	}
	foreignOnly := []LookupMatch{m("5", "STOCK", "JP", "1234.T-JP")}

	tests := []struct {
		name       string
		query      string
		matches    []LookupMatch
		want       string
		wantLondon []string
	}{
		{"US stock over ETF", "anthropic", anthropic, "ANT", nil},
		{"exact ticker wins, any case", "antw", anthropic, "ANTW", nil},
		{"exact US ticker of a UK company", "bp", bp, "BP", []string{"BP.-GB"}},
		{"exact London ticker", "bp.-gb", bp, "BP.-GB", nil},
		{"multi-word query", "anthropic+pbc", anthropic, "ANT", nil},
		{"UK company: US line, London to check", "tesco", tesco, "TSCDF", []string{"TSCO-GB"}},
		{"US company with a London line", "apple", apple, "AAPL", []string{"0R2V-GB"}},
		{"UK company ranked by its ADR", "bp+plc", bp, "BP", []string{"BP.-GB"}},
		{"other foreign company: US line", "toyota", toyota, "TM", nil},
		{"other foreign company ranked at home", "nestle", nestle, "NSRGY", nil},
		{"London-only company beats foreign", "abc", ukOnly, "ABC-GB", []string{"ABC-GB"}},
		{"name match beats CNBC's ranking", "astrazeneca", astra, "AZN", []string{"AZN-GB"}},
		{"whole-word name match", "novo", novo, "NVO", nil},
		{"no stocks: US listing", "ant+etf", etfs, "ANTC", nil},
		{"foreign only", "x", foreignOnly, "1234.T-JP", nil},
		{"metadata only", "zzz", anthropic[:1], "", nil},
		{"no matches", "zzz", nil, "", nil},
	}
	for _, tt := range tests {
		got, london := pickSymbol(tt.query, tt.matches)
		if got != tt.want || strings.Join(london, ",") != strings.Join(tt.wantLondon, ",") {
			t.Errorf("%s: pickSymbol(%q) = %q, %q; want %q, %q", tt.name, tt.query, got, london, tt.want, tt.wantLondon)
		}
	}
}

func TestNameStartsWith(t *testing.T) {
	tests := []struct {
		name, query string
		want        bool
	}{
		{"Rolls-Royce Holdings PLC", "rolls royce", true},
		{"Novo Nordisk A/S", "novo", true},
		{"NOVONIX Ltd", "novo", false},
		{"Anthropic, PBC", "anthropic", true},
		{"Apple", "apple", true},
		{"Gilead Sciences, Inc.", "astrazeneca", false},
		{"Anything", "", false},
	}
	for _, tt := range tests {
		if got := nameStartsWith(tt.name, tt.query); got != tt.want {
			t.Errorf("nameStartsWith(%q, %q) = %v, want %v", tt.name, tt.query, got, tt.want)
		}
	}
}

func TestUKListing(t *testing.T) {
	tests := []struct {
		name   string
		quotes []*stock
		want   string
	}{
		{"pence", []*stock{{symbol: "TSCO-GB", currency: "GBp", price: "548.80"}}, "TSCO-GB"},
		{"foreign line in USD", []*stock{{symbol: "0R2V-GB", currency: "USD"}}, ""},
		{"no price", []*stock{{symbol: "X-GB", currency: "GBp"}}, ""},
		{"first priced wins", []*stock{{symbol: "A-GB", currency: "GBX"}, {symbol: "B-GB", currency: "GBP", price: "1"}}, "B-GB"},
		{"none", nil, ""},
	}
	for _, tt := range tests {
		if got := ukListing(tt.quotes); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
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

func TestPriceUnit(t *testing.T) {
	for currency, want := range map[string]string{"GBp": "p", "GBX": "p", "GBP": "", "USD": "", "": ""} {
		if got := priceUnit(currency); got != want {
			t.Errorf("priceUnit(%q) = %q, want %q", currency, got, want)
		}
	}
	if got := withUnit("548.60", "p"); got != "548.60p" {
		t.Errorf("withUnit = %q", got)
	}
	if got := withUnit("", "p"); got != "" {
		t.Errorf("withUnit empty = %q", got)
	}
}

func TestAddAnsiColoursLondonSymbol(t *testing.T) {
	got := addAnsiColours("BA.-GB   1,828.00p  +14.00  (+0.77%)\nTSCO-GB  546.40p  -6.60  (-1.19%)")
	if !strings.Contains(got, "\x1b[2;32m+14.00\x1b[0m") || !strings.Contains(got, "\x1b[2;32m+0.77%\x1b[0m") {
		t.Errorf("rise on a hyphenated symbol not green: %q", got)
	}
	if !strings.Contains(got, "\x1b[2;31m-6.60\x1b[0m") || strings.Contains(got, "\x1b[2;31m-GB") {
		t.Errorf("fall colouring wrong: %q", got)
	}
}
