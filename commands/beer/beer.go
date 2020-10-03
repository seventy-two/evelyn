package beer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
	"github.com/seventy-two/evelyn/service"
)

var serviceConfig *service.Service

type beerResult struct {
	Meta struct {
		Code         int `json:"code"`
		ResponseTime struct {
			Time    float64 `json:"time"`
			Measure string  `json:"measure"`
		} `json:"response_time"`
		InitTime struct {
			Time    int    `json:"time"`
			Measure string `json:"measure"`
		} `json:"init_time"`
	} `json:"meta"`
	Response struct {
		Message       string  `json:"message"`
		TimeTaken     float64 `json:"time_taken"`
		BreweryID     int     `json:"brewery_id"`
		SearchType    string  `json:"search_type"`
		TypeID        int     `json:"type_id"`
		SearchVersion int     `json:"search_version"`
		Found         int     `json:"found"`
		Offset        int     `json:"offset"`
		Limit         int     `json:"limit"`
		Term          string  `json:"term"`
		ParsedTerm    string  `json:"parsed_term"`
		Beers         struct {
			Count int `json:"count"`
			Items []struct {
				CheckinCount int  `json:"checkin_count"`
				HaveHad      bool `json:"have_had"`
				YourCount    int  `json:"your_count"`
				Beer         struct {
					Bid             int     `json:"bid"`
					BeerName        string  `json:"beer_name"`
					BeerLabel       string  `json:"beer_label"`
					BeerAbv         float64 `json:"beer_abv"`
					BeerSlug        string  `json:"beer_slug"`
					BeerIbu         int     `json:"beer_ibu"`
					BeerDescription string  `json:"beer_description"`
					CreatedAt       string  `json:"created_at"`
					BeerStyle       string  `json:"beer_style"`
					InProduction    int     `json:"in_production"`
					AuthRating      float32 `json:"auth_rating"`
					WishList        bool    `json:"wish_list"`
				} `json:"beer"`
				Brewery struct {
					BreweryID      int    `json:"brewery_id"`
					BreweryName    string `json:"brewery_name"`
					BrewerySlug    string `json:"brewery_slug"`
					BreweryPageURL string `json:"brewery_page_url"`
					BreweryType    string `json:"brewery_type"`
					BreweryLabel   string `json:"brewery_label"`
					CountryName    string `json:"country_name"`
					Contact        struct {
						Twitter   string `json:"twitter"`
						Facebook  string `json:"facebook"`
						Instagram string `json:"instagram"`
						URL       string `json:"url"`
					} `json:"contact"`
					Location struct {
						BreweryCity  string  `json:"brewery_city"`
						BreweryState string  `json:"brewery_state"`
						Lat          float64 `json:"lat"`
						Lng          float64 `json:"lng"`
					} `json:"location"`
					BreweryActive int `json:"brewery_active"`
				} `json:"brewery"`
			} `json:"items"`
		} `json:"beers"`
		Homebrew struct {
			Count int `json:"count"`
			Items []struct {
				CheckinCount int  `json:"checkin_count"`
				HaveHad      bool `json:"have_had"`
				YourCount    int  `json:"your_count"`
				Beer         struct {
					Bid             int     `json:"bid"`
					BeerName        string  `json:"beer_name"`
					BeerLabel       string  `json:"beer_label"`
					BeerAbv         float64 `json:"beer_abv"`
					BeerSlug        string  `json:"beer_slug"`
					BeerIbu         int     `json:"beer_ibu"`
					BeerDescription string  `json:"beer_description"`
					CreatedAt       string  `json:"created_at"`
					BeerStyle       string  `json:"beer_style"`
					InProduction    int     `json:"in_production"`
					AuthRating      float32 `json:"auth_rating"`
					WishList        bool    `json:"wish_list"`
				} `json:"beer"`
				Brewery struct {
					BreweryID      int    `json:"brewery_id"`
					BreweryName    string `json:"brewery_name"`
					BrewerySlug    string `json:"brewery_slug"`
					BreweryPageURL string `json:"brewery_page_url"`
					BreweryType    string `json:"brewery_type"`
					BreweryLabel   string `json:"brewery_label"`
					CountryName    string `json:"country_name"`
					Contact        struct {
						Twitter   string `json:"twitter"`
						Facebook  string `json:"facebook"`
						Instagram string `json:"instagram"`
						URL       string `json:"url"`
					} `json:"contact"`
					Location struct {
						BreweryCity  string  `json:"brewery_city"`
						BreweryState string  `json:"brewery_state"`
						Lat          float64 `json:"lat"`
						Lng          float64 `json:"lng"`
					} `json:"location"`
					BreweryActive int `json:"brewery_active"`
				} `json:"brewery"`
			} `json:"items"`
		} `json:"homebrew"`
		Breweries struct {
			Items []interface{} `json:"items"`
			Count int           `json:"count"`
		} `json:"breweries"`
	} `json:"response"`
}

type beer struct {
	name        string
	label       string
	abv         string
	desc        string
	style       string
	brewery     string
	breweryType string
	country     string
	rating      string
}

func getBeers(query string) ([]*beer, error) {
	results := &beerResult{}
	err := web.GetJSON(fmt.Sprintf(serviceConfig.TargetURL, url.QueryEscape(query), serviceConfig.APIKey), results)
	if err != nil {
		return nil, err
	}
	var rBeers []*beer
	for _, b := range results.Response.Beers.Items {
		rBeers = append(rBeers, &beer{
			name:        b.Beer.BeerName,
			label:       b.Beer.BeerLabel,
			abv:         fmt.Sprintf("%.2f", b.Beer.BeerAbv),
			desc:        b.Beer.BeerDescription,
			style:       b.Beer.BeerStyle,
			brewery:     b.Brewery.BreweryName,
			breweryType: b.Brewery.BreweryType,
			country:     b.Brewery.CountryName,
			rating:      getRating(b.Beer.AuthRating),
		})
	}
	for _, b := range results.Response.Homebrew.Items {
		rBeers = append(rBeers, &beer{
			name:        b.Beer.BeerName,
			label:       b.Beer.BeerLabel,
			abv:         fmt.Sprintf("%.2f", b.Beer.BeerAbv),
			desc:        b.Beer.BeerDescription,
			style:       b.Beer.BeerStyle,
			brewery:     b.Brewery.BreweryName,
			breweryType: b.Brewery.BreweryType,
			country:     b.Brewery.CountryName,
			rating:      getRating(b.Beer.AuthRating),
		})
	}
	return rBeers, nil
}

func getRating(rating float32) string {
	if rating != 0 {
		// rating = rating * 2
		// str := ""
		// r := int(math.Round(float64(rating)))
		// for i := 0; i < r; i++ {
		// 	str += "🍺"
		// }
		// return str
		rating = rating * 20
		return fmt.Sprintf("%.0f%s", rating, "%")
	}
	return "Not Yet Rated"
}

// RegisterService registers the beer service
func RegisterService(dg *discordgo.Session, config *service.Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")

	switch matches[0] {
	case "!beer":
		beers, err := getBeers(strings.Join(matches[1:], "+"))
		if err != nil {
			str := fmt.Sprintf("an error occured (%s)", err)
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
			return
		}
		if len(beers) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No beers :(")
			return
		}
		firstBeer := beers[0]
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       firstBeer.name,
			Description: firstBeer.desc,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: firstBeer.label,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "ABV",
					Value:  firstBeer.abv + "%",
					Inline: true,
				},
				{
					Name:   "Style",
					Value:  firstBeer.style,
					Inline: true,
				},
				{
					Name:   "Brewery",
					Value:  firstBeer.brewery,
					Inline: true,
				},
				{
					Name:   "Brewery Type",
					Value:  firstBeer.breweryType,
					Inline: true,
				},
				{
					Name:   "Country",
					Value:  firstBeer.country,
					Inline: true,
				},
				{
					Name:   "72's Rating",
					Value:  firstBeer.rating,
					Inline: true,
				},
			},
		})
	}
}
