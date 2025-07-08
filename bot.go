package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/seventy-two/evelyn/database"

	"github.com/seventy-two/evelyn/commands/beer"
	"github.com/seventy-two/evelyn/commands/bing"
	"github.com/seventy-two/evelyn/commands/olympics"
	"github.com/seventy-two/evelyn/commands/stocks"
	"github.com/seventy-two/evelyn/passive/generation"
	"github.com/seventy-two/evelyn/passive/images"
	"github.com/seventy-two/evelyn/passive/shitpost"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	cli "github.com/jawher/mow.cli"
	"github.com/seventy-two/evelyn/commands/dictionary"
	"github.com/seventy-two/evelyn/commands/dota"
	"github.com/seventy-two/evelyn/commands/math"
	"github.com/seventy-two/evelyn/commands/movie"
	"github.com/seventy-two/evelyn/commands/nfl"
	"github.com/seventy-two/evelyn/commands/quotes"
	"github.com/seventy-two/evelyn/commands/siege"
	"github.com/seventy-two/evelyn/commands/tv"
	"github.com/seventy-two/evelyn/commands/urbandictionary"
	"github.com/seventy-two/evelyn/commands/weather"
)

func startedUp(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateCustomStatus("Listening to ! prefix")
}

func start(app *cli.Cli, services *serviceConfig, dbPath string, logFile string, errorLog string) {
	dg, _ := discordgo.New(fmt.Sprintf("Bot %s", services.discordAPI.APIKey))
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	go registerServices(dg, services, dbPath)
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

	dg.AddHandler(createLogger(logFile))

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func registerServices(dg *discordgo.Session, services *serviceConfig, dbPath string) {

	db := database.NewDatabase(dbPath)

	if services.dictionaryAPI != nil {
		dictionary.RegisterService(dg, services.dictionaryAPI)
	}
	if services.dotaAPI != nil {
		dota.RegisterService(dg, services.dotaAPI)
	}
	if services.nflAPI != nil {
		nfl.RegisterService(dg, services.nflAPI)
	}
	if services.movieAPI != nil {
		movie.RegisterService(dg, services.movieAPI)
	}
	if services.stocksAPI != nil {
		services.stocksAPI.Storage = stocks.NewWatchlistStorage(db)
		stocks.RegisterService(dg, services.stocksAPI)
	}
	if services.tvAPI != nil {
		tv.RegisterService(dg, services.tvAPI)
	}
	if services.urbanAPI != nil {
		urbandictionary.RegisterService(dg, services.urbanAPI)
	}
	if services.weatherAPI != nil {
		services.weatherAPI.LocationStorage = &weather.LocationStorage{Db: db}
		weather.RegisterService(dg, services.weatherAPI)
	}
	if services.mathAPI != nil {
		math.RegisterService(dg, services.mathAPI)
	}
	if services.siegeAPI != nil {
		siege.RegisterService(dg, services.siegeAPI)
	}
	if services.beerAPI != nil {
		beer.RegisterService(dg, services.beerAPI)
	}
	if services.quotesAPI != nil {
		quotes.RegisterService(dg, services.quotesAPI)
	}
	if services.bingAPI != nil {
		bing.RegisterService(dg, services.bingAPI)
	}
	if services.olympicsAPI != nil {
		olympics.RegisterService(dg, services.olympicsAPI)
	}
	if services.generationAPI != nil {
		generation.RegisterService(dg, services.generationAPI)
	}

	shitpost.RegisterService(dg, &shitpost.Service{Db: db})
	images.RegisterService(dg, &images.Service{})
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
