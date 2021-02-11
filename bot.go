package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/seventy-two/evelyn/database"

	"github.com/seventy-two/evelyn/commands/beer"
	"github.com/seventy-two/evelyn/commands/shitpost"

	"github.com/bwmarrin/discordgo"
	cli "github.com/jawher/mow.cli"
	"github.com/seventy-two/evelyn/commands/dictionary"
	"github.com/seventy-two/evelyn/commands/dota"
	"github.com/seventy-two/evelyn/commands/math"
	"github.com/seventy-two/evelyn/commands/movie"
	"github.com/seventy-two/evelyn/commands/nfl"
	"github.com/seventy-two/evelyn/commands/quotes"
	"github.com/seventy-two/evelyn/commands/siege"
	"github.com/seventy-two/evelyn/commands/stocks"
	"github.com/seventy-two/evelyn/commands/tv"
	"github.com/seventy-two/evelyn/commands/urbandictionary"
	"github.com/seventy-two/evelyn/commands/weather"
	"github.com/seventy-two/evelyn/passive/upsetter"
)

func startedUp(s *discordgo.Session, event *discordgo.Ready) {
	s.UserUpdateStatus(discordgo.Status("Listening to ! prefix"))
}

func start(app *cli.Cli, services *serviceConfig, dbPath string) {
	dg, _ := discordgo.New(fmt.Sprintf("Bot %s", services.discordAPI.APIKey))
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	go registerServices(dg, services, dbPath)
	dg.AddHandler(startedUp)
	err := dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %s", err)
	}

	dg.AddHandler(logger)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
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

	shitpost.RegisterService(dg)

	go upsetter.Upset(dg)
}

func logger(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Author.ID + " | " + m.Author.Username + " | " + m.Content)
}
