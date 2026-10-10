package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/seventy-two/evelyn/database"

	"github.com/seventy-two/evelyn/commands/beer"
	"github.com/seventy-two/evelyn/commands/bing"
	"github.com/seventy-two/evelyn/commands/help"
	"github.com/seventy-two/evelyn/commands/stocks"
	"github.com/seventy-two/evelyn/passive/autotrader"
	"github.com/seventy-two/evelyn/passive/generation"
	"github.com/seventy-two/evelyn/passive/images"
	"github.com/seventy-two/evelyn/passive/shitpost"
	"github.com/seventy-two/evelyn/slash"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	cli "github.com/jawher/mow.cli"
	"github.com/seventy-two/evelyn/commands/dictionary"
	"github.com/seventy-two/evelyn/commands/math"
	"github.com/seventy-two/evelyn/commands/movie"
	"github.com/seventy-two/evelyn/commands/nfl"
	"github.com/seventy-two/evelyn/commands/quotes"
	"github.com/seventy-two/evelyn/commands/tv"
	"github.com/seventy-two/evelyn/commands/urbandictionary"
	"github.com/seventy-two/evelyn/commands/weather"
)

func startedUp(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateCustomStatus("Listening to ! prefix and /commands")
}

func start(app *cli.Cli, services *serviceConfig, dbPath string, logFile string, errorLog string) {
	dg, _ := discordgo.New(fmt.Sprintf("Bot %s", services.discordAPI.APIKey))
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	logger := createLogger(logFile)
	go registerServices(dg, services, dbPath, logger)
	dg.AddHandler(startedUp)
	err := dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %s", err)
	}

	errLog, err := os.OpenFile(errorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(errLog)

	dg.AddHandler(logger)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func registerServices(dg *discordgo.Session, services *serviceConfig, dbPath string, logger func(*discordgo.Session, *discordgo.MessageCreate)) {

	db := database.NewDatabase(dbPath)
	router := slash.NewRouter(logger)

	if services.dictionaryAPI != nil {
		dictionary.RegisterService(dg, services.dictionaryAPI)
		router.Add(dictionary.SlashCommands()...)
	}
	if services.nflAPI != nil {
		nfl.RegisterService(dg, services.nflAPI)
		router.Add(nfl.SlashCommands()...)
	}
	if services.movieAPI != nil {
		movie.RegisterService(dg, services.movieAPI)
		router.Add(movie.SlashCommands()...)
	}
	if services.stocksAPI != nil {
		services.stocksAPI.Storage = stocks.NewWatchlistStorage(db)
		stocks.RegisterService(dg, services.stocksAPI)
		router.Add(stocks.SlashCommands()...)
	}
	if services.tvAPI != nil {
		tv.RegisterService(dg, services.tvAPI)
		router.Add(tv.SlashCommands()...)
	}
	if services.urbanAPI != nil {
		urbandictionary.RegisterService(dg, services.urbanAPI)
		router.Add(urbandictionary.SlashCommands()...)
	}
	if services.weatherAPI != nil {
		services.weatherAPI.LocationStorage = &weather.LocationStorage{Db: db}
		weather.RegisterService(dg, services.weatherAPI)
		router.Add(weather.SlashCommands()...)
	}
	if services.mathAPI != nil {
		math.RegisterService(dg, services.mathAPI)
		router.Add(math.SlashCommands()...)
	}
	if services.beerAPI != nil {
		beer.RegisterService(dg, services.beerAPI)
		router.Add(beer.SlashCommands()...)
	}
	if services.quotesAPI != nil {
		quotes.RegisterService(dg, services.quotesAPI)
		router.Add(quotes.SlashCommands()...)
	}
	if services.bingAPI != nil {
		bing.RegisterService(dg, services.bingAPI)
	}
	if services.generationAPI != nil {
		generation.RegisterService(dg, services.generationAPI)
		router.Add(generation.SlashCommands()...)
	}

	if services.autotraderAPI != nil {
		services.autotraderAPI.Store = db
		autotrader.RegisterService(dg, services.autotraderAPI)
	}

	shitpost.RegisterService(dg, &shitpost.Service{Db: db})
	images.RegisterService(dg, &images.Service{})
	help.RegisterService(dg)
	router.Add(help.SlashCommands()...)

	if err := router.Register(dg); err != nil {
		log.Errorf("slash commands unavailable: %s", err)
	}
}

func createLogger(logfile string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		content := fmt.Sprintf("[%s] %s (%s): %s", m.Timestamp, m.Author.Username, m.Author.ID, m.Content)
		fmt.Fprintln(f, content)
	}
}
