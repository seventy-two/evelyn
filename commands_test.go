package main

import (
	"regexp"
	"testing"

	"github.com/seventy-two/evelyn/commands/beer"
	"github.com/seventy-two/evelyn/commands/dictionary"
	"github.com/seventy-two/evelyn/commands/dota"
	"github.com/seventy-two/evelyn/commands/help"
	"github.com/seventy-two/evelyn/commands/math"
	"github.com/seventy-two/evelyn/commands/movie"
	"github.com/seventy-two/evelyn/commands/nfl"
	"github.com/seventy-two/evelyn/commands/olympics"
	"github.com/seventy-two/evelyn/commands/quotes"
	"github.com/seventy-two/evelyn/commands/siege"
	"github.com/seventy-two/evelyn/commands/stocks"
	"github.com/seventy-two/evelyn/commands/tv"
	"github.com/seventy-two/evelyn/commands/urbandictionary"
	"github.com/seventy-two/evelyn/commands/wc"
	"github.com/seventy-two/evelyn/commands/weather"
	"github.com/seventy-two/evelyn/passive/generation"
	"github.com/seventy-two/evelyn/slash"
)

// TestSlashCommandsValid checks every slash command against Discord's rules,
// since one invalid definition makes Discord reject the whole set.
func TestSlashCommandsValid(t *testing.T) {
	name := regexp.MustCompile(`^[-_a-z0-9]{1,32}$`)
	var all []*slash.Command
	for _, cmds := range [][]*slash.Command{
		beer.SlashCommands(), dictionary.SlashCommands(), dota.SlashCommands(), help.SlashCommands(),
		math.SlashCommands(), movie.SlashCommands(), nfl.SlashCommands(), olympics.SlashCommands(),
		quotes.SlashCommands(), siege.SlashCommands(), stocks.SlashCommands(), tv.SlashCommands(),
		urbandictionary.SlashCommands(), wc.SlashCommands(), weather.SlashCommands(), generation.SlashCommands(),
	} {
		all = append(all, cmds...)
	}

	seen := map[string]bool{}
	for _, c := range all {
		if !name.MatchString(c.Name) || seen[c.Name] {
			t.Errorf("bad or duplicate name %q", c.Name)
		}
		seen[c.Name] = true
		if l := len(c.Description); l < 1 || l > 100 {
			t.Errorf("/%s: description length %d", c.Name, l)
		}
		if c.Handler == nil {
			t.Errorf("/%s: no handler", c.Name)
		}
		optional := false
		for _, o := range c.Options {
			if !name.MatchString(o.Name) || len(o.Description) < 1 || len(o.Description) > 100 {
				t.Errorf("/%s: bad option %q", c.Name, o.Name)
			}
			if o.Required && optional {
				t.Errorf("/%s: required option %q after optional one", c.Name, o.Name)
			}
			optional = optional || !o.Required
			for _, ch := range o.Choices {
				if len(ch.Name) < 1 || len(ch.Name) > 100 {
					t.Errorf("/%s %s: bad choice %q", c.Name, o.Name, ch.Name)
				}
			}
		}
	}
	t.Logf("%d slash commands", len(all))
}
