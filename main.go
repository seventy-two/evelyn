package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/seventy-two/evelyn/commands/dictionary"
	"github.com/seventy-two/evelyn/commands/stocks"
	"github.com/seventy-two/evelyn/commands/weather"
	"github.com/seventy-two/evelyn/passive/autotrader"
	"github.com/seventy-two/evelyn/passive/generation"
	"github.com/seventy-two/evelyn/service"

	cli "github.com/jawher/mow.cli"
)

type appLink struct {
	name string
	url  string
}

type serviceConfig struct {
	discordAPI    *service.Service
	weatherAPI    *weather.Service
	divegrassAPI  *service.Service
	dictionaryAPI *dictionary.Service
	movieAPI      *service.Service
	mathAPI       *service.Service
	tvAPI         *service.Service
	urbanAPI      *service.Service
	nflAPI        *service.Service
	stocksAPI     *stocks.Service
	beerAPI       *service.Service
	quotesAPI     *service.Service
	bingAPI       *service.Service
	generationAPI *generation.Service
	autotraderAPI *autotrader.Service
}

var appMeta = struct {
	name        string
	description string
	discord     string
	maintainers string
	links       []appLink
}{
	name:        "Evelyn",
	description: "Discord assistant with various commands",
	discord:     "https://discord.gg/F2cD4cN",
	maintainers: "github.com/seventy-two",
	links: []appLink{
		{name: "vcs", url: "https://github.com/seventy-two/evelyn"},
	},
}

func main() {

	app := cli.App(appMeta.name, appMeta.description)

	Services := &serviceConfig{
		discordAPI: &service.Service{
			APIKey: *app.String(cli.StringOpt{
				Name:   "EvelynDiscordAPIKey",
				Value:  "",
				EnvVar: "EVELYN_DISCORD_API_KEY",
			}),
		},
		weatherAPI: &weather.Service{
			GeoCodeURL: *app.String(cli.StringOpt{
				Name:   "GeoCodeURL",
				Value:  "https://maps.googleapis.com/maps/api/geocode/json?address=%s&region=UK&key=%s",
				EnvVar: "GEOCODE_URL",
			}),
			GeoCodeAPIKey: *app.String(cli.StringOpt{
				Name:   "GeoCodeAPIKey",
				Value:  "",
				EnvVar: "GEOCODE_API_KEY",
			}),
			DarkSkyURL: *app.String(cli.StringOpt{
				Name:   "WeatherURL",
				Value:  "https://api.pirateweather.net/forecast/%s/%s?units=uk&exclude=minutely,hourly,alerts",
				EnvVar: "WEATHER_URL",
			}),
			DarkSkyAPIKey: *app.String(cli.StringOpt{
				Name:   "WeatherAPIKey",
				Value:  "",
				EnvVar: "WEATHER_API_KEY",
			}),
		},
		dictionaryAPI: &dictionary.Service{
			WordnikSearchURL: *app.String(cli.StringOpt{
				Name:   "WordnikSearchURL",
				Value:  "http://api.wordnik.com/v4/word.json/%s/definitions?limit=3&includeRelated=true&sourceDictionaries=all&useCanonical=true&includeTags=false&api_key=%s",
				EnvVar: "WORDNIK_SEARCH_URL",
			}),
			WordnikRelationshipURL: *app.String(cli.StringOpt{
				Name:   "WordnikRelationshipURL",
				Value:  "https://api.wordnik.com/v4/word.json/%s/relatedWords?useCanonical=false&relationshipTypes=equivalent&limitPerRelationshipType=500&api_key=%s",
				EnvVar: "WORDNIK_RELATIONSHIP_URL",
			}),
			WOTDURL: *app.String(cli.StringOpt{
				Name:   "WOTDURL",
				Value:  "http://api.wordnik.com:80/v4/words.json/wordOfTheDay?api_key=%s",
				EnvVar: "WOTD_URL",
			}),
			WordnikAPIKey: *app.String(cli.StringOpt{
				Name:   "WordnikAPIKey",
				Value:  "",
				EnvVar: "WORDNIK_API_KEY",
			}),
		},
		movieAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "MovieURL",
				Value:  "http://www.omdbapi.com/?t=%s&plot=short&apikey=%s%s",
				EnvVar: "MOVIE_URL",
			}),
			APIKey: *app.String(cli.StringOpt{
				Name:   "MovieAPIKey",
				Value:  "",
				EnvVar: "MOVIE_API_KEY",
			}),
		},
		divegrassAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "divegrassURL",
				Value:  "https://api.fifa.com/api/v1/live/football/",
				EnvVar: "DIVEGRASS_URL",
			}),
		},
		mathAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "mathURL",
				Value:  "http://api.wolframalpha.com/v2/query?appid=%s&input=%s",
				EnvVar: "MATH_URL",
			}),
			APIKey: *app.String(cli.StringOpt{
				Name:   "mathAPIKey",
				Value:  "",
				EnvVar: "MATH_API_KEY",
			}),
		},
		tvAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "TvURL",
				Value:  "http://api.tvmaze.com/singlesearch/shows?q=%s",
				EnvVar: "TV_URL",
			}),
		},
		nflAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "NFLURL",
				Value:  "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard",
				EnvVar: "NFL_API_URL",
			}),
		},
		stocksAPI: &stocks.Service{
			QuoteURL: *app.String(cli.StringOpt{
				Name:   "StocksQuoteURL",
				Value:  "https://quote.cnbc.com/quote-html-webservice/restQuote/symbolType/symbol?symbols=%s&requestMethod=itv&noform=1&partnerId=2&fund=1&exthrs=1&output=json&events=1",
				EnvVar: "STOCKS_QUOTE_URL",
			}),
			MoverURL: *app.String(cli.StringOpt{
				Name:   "StocksMoverURL",
				Value:  "https://gdsapi.cnbc.com/market-mover/groupMover/%s/CHANGE_PCT/BOTH/5.json?source=SAVED&delayed=false&partnerId=2",
				EnvVar: "STOCKS_MOVER_URL",
			}),
			APIKey: *app.String(cli.StringOpt{
				Name:   "StocksQuoteKey",
				Value:  "",
				EnvVar: "STOCKS_QUOTE_KEY",
			}),
			LookupURL: *app.String(cli.StringOpt{
				Name:   "StocksLookupURL",
				Value:  "https://www.alphavantage.co/query?function=SYMBOL_SEARCH&keywords=%s&apikey=%s",
				EnvVar: "STOCKS_LOOKUP_URL",
			}),
			EarningsURL: *app.String(cli.StringOpt{
				Name:   "StocksEarningsURL",
				Value:  "https://www.alphavantage.co/query?function=EARNINGS_CALENDAR&horizon=1week&apikey=%s",
				EnvVar: "STOCKS_EARNINGS_URL",
			}),
			ExchangeURL: *app.String(cli.StringOpt{
				Name:   "StocksExchangeURL",
				Value:  "https://www.alphavantage.co/query?function=CURRENCY_EXCHANGE_RATE&from_currency=%s&to_currency=%s&apikey=%s",
				EnvVar: "STOCKS_EXCHANGE_URL",
			}),
			LogoURL: *app.String(cli.StringOpt{
				Name:   "StocksLogoURL",
				Value:  "https://img.logo.dev/ticker/%s?token=%s",
				EnvVar: "STOCKS_LOGO_URL",
			}),
			LogoAPIKey: *app.String(cli.StringOpt{
				Name:   "StocksLogoAPIKey",
				Value:  "",
				EnvVar: "STOCKS_LOGO_API_KEY",
			}),
		},
		urbanAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "UrbanAPIURL",
				Value:  "http://api.urbandictionary.com/v0/define?term=%s",
				EnvVar: "URBAN_DICT_URL",
			}),
		},
		beerAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "BeerAPIURL",
				Value:  "https://api.untappd.com/v4/search/beer?q=%s&%s",
				EnvVar: "BEER_API_URL",
			}),
			APIKey: *app.String(cli.StringOpt{
				Name:   "BeerAPIKey",
				Value:  "",
				EnvVar: "BEER_API_KEY",
			}),
		},
		quotesAPI: &service.Service{
			TargetURL: *app.String(cli.StringOpt{
				Name:   "QuotesAPIURL",
				Value:  "https://zenquotes.io/api/today",
				EnvVar: "QUOTES_API_URL",
			}),
		},
		bingAPI: &service.Service{
			APIKey: *app.String(cli.StringOpt{
				Name:   "BingAPIKey",
				Value:  "",
				EnvVar: "BING_API_KEY",
			}),
			TargetURL: *app.String(cli.StringOpt{
				Name:   "BingAPIURL",
				Value:  "https://api.bing.microsoft.com/v7.0/images/search?q=",
				EnvVar: "BING_API_URL",
			}),
		},
		generationAPI: &generation.Service{
			APIKey: *app.String(cli.StringOpt{
				Name:   "GenerationAPIKey",
				Value:  "",
				EnvVar: "GENERATION_API_KEY",
			}),
			YoutubeURL: *app.String(cli.StringOpt{
				Name:   "GenerationYoutubeURL",
				Value:  "https://www.googleapis.com/youtube/v3/videos?id=%s&part=contentDetails&key=%s",
				EnvVar: "GENERATION_YOUTUBE_URL",
			}),
			YoutubeAPIKey: *app.String(cli.StringOpt{
				Name:   "GenerationYoutubeAPIKey",
				Value:  "",
				EnvVar: "GENERATION_YOUTUBE_API_KEY",
			}),
		},
	}

	Services.autotraderAPI = &autotrader.Service{
		SearchURL: *app.String(cli.StringOpt{
			Name:   "AutotraderSearchURL",
			Value:  "",
			EnvVar: "AUTOTRADER_SEARCH_URL",
		}),
		ChannelID: *app.String(cli.StringOpt{
			Name:   "AutotraderChannelID",
			Value:  "237647756049514498",
			EnvVar: "AUTOTRADER_CHANNEL_ID",
		}),
		UserID: *app.String(cli.StringOpt{
			Name:   "AutotraderUserID",
			Value:  "176722208243187712",
			EnvVar: "AUTOTRADER_USER_ID",
		}),
	}
	autotraderInterval := *app.String(cli.StringOpt{
		Name:   "AutotraderInterval",
		Value:  "2h",
		EnvVar: "AUTOTRADER_INTERVAL",
	})

	dbPath := *app.String(cli.StringOpt{
		Name:   "DBPath",
		Value:  "/root/evelynDB",
		EnvVar: "DB_PATH",
	})

	logPath := *app.String(cli.StringOpt{
		Name:   "LogPath",
		Value:  "/root/evelyn.new.log",
		EnvVar: "LOG_PATH",
	})

	errPath := *app.String(cli.StringOpt{
		Name:   "ErrPath",
		Value:  "/root/evelyn.err",
		EnvVar: "ERR_PATH",
	})

	app.Action = func() {
		interval, err := time.ParseDuration(autotraderInterval)
		if err != nil || interval < 10*time.Minute {
			log.Errorf("AutoTrader watcher off: AUTOTRADER_INTERVAL must be a duration of at least 10m, got %q", autotraderInterval)
			Services.autotraderAPI.SearchURL = ""
		}
		Services.autotraderAPI.Interval = interval
		start(app, Services, dbPath, logPath, errPath)
	}

	app.Run(os.Args)
}
