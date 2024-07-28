package weather

import (
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/seventy-two/Cara/web"
)

var serviceConfig *Service

type Service struct {
	GeoCodeAPIKey   string
	GeoCodeURL      string
	DarkSkyAPIKey   string
	DarkSkyURL      string
	LocationStorage *LocationStorage
}

func emoji(icon string) string {
	if icon == "clear-day" {
		return "☀️"
	} else if icon == "clear-night" {
		return "🌙"
	} else if icon == "rain" {
		return "☔️"
	} else if icon == "snow" {
		return "❄️"
	} else if icon == "sleet" {
		return "☔️❄️"
	} else if icon == "wind" {
		return "💨"
	} else if icon == "fog" {
		return "🌁"
	} else if icon == "cloudy" {
		return "☁️"
	} else if icon == "partly-cloudy-day" {
		return "⛅"
	} else if icon == "partly-cloudy-night" {
		return "⛅"
	} else {
		return ""
	}
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func getCoords(location string) string {
	var err error
	geo := &geocodeResponse{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.GeoCodeURL, url.QueryEscape(location), serviceConfig.GeoCodeAPIKey), geo)
	if err != nil || geo.Status != "OK" {
		return ""
	}
	return fmt.Sprintf("%v,%v", geo.Results[0].Geometry.Location.Lat, geo.Results[0].Geometry.Location.Lng)
}

func weather(location string) (msg string, err error) {
	location = strings.Title(location)
	coords := getCoords(location)
	if coords == "" {
		return fmt.Sprintf("Could not find %s", location), nil
	}

	data := &forecastResponse{}
	err = web.GetJSON(fmt.Sprintf(serviceConfig.DarkSkyURL, serviceConfig.DarkSkyAPIKey, coords), data)
	if err != nil {
		return fmt.Sprintf("Could not get weather for: %s", location), nil
	}

	units := "°C"
	windspeed := "mph"

	return fmt.Sprintf("%s (%s)\nNow: %s %s %v%s\nToday: %s %v%s/%v%s\nHumidity: %v%% Wind: %v%s Precipitation: %v%%",
		location,
		coords,
		data.Currently.Summary,
		emoji(data.Currently.Icon),
		round(data.Currently.Temperature),
		units,
		emoji(data.Daily.Data[0].Icon),
		round(data.Daily.Data[0].TemperatureMax),
		units,
		round(data.Daily.Data[0].TemperatureMin),
		units,
		int(data.Daily.Data[0].Humidity*100),
		data.Daily.Data[0].WindSpeed,
		windspeed,
		int(data.Daily.Data[0].PrecipProbability*100)), nil
}

func RegisterService(dg *discordgo.Session, config *Service) {
	serviceConfig = config
	dg.AddHandler(invokeCommand)
}

func invokeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	matches := strings.Split(m.Content, " ")
	switch matches[0] {
	case "!w":
		var location string
		var err error
		if len(matches) < 2 {
			location, err = serviceConfig.LocationStorage.getLocation(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Set your location using !set location <location>")
				return
			}
		} else {
			location = strings.Join(matches[1:], " ")
		}
		str, err := weather(location)
		if err != nil {
			fmt.Println(fmt.Sprintf("an error occured (%s)", err.Error()))
			return
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}

	case "!f":
		var location string
		var err error
		if len(matches) < 2 {
			location, err = serviceConfig.LocationStorage.getLocation(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Set your location using !set location <location>")
				return
			}
		} else {
			location = strings.Join(matches[1:], " ")
		}
		str, err := forecast(location)
		if err != nil {
			fmt.Printf("an error occured (%s)", err.Error())
			return
		}

		if str != "" {
			fmtstr := fmt.Sprintf("```%s```", str)
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	case "!set":
		if len(matches) < 2 {
			s.ChannelMessageSend(m.ChannelID, "stfu u stupid fuckin retard")
			return
		}
		if matches[1] != "location" {
			return
		}
		l := matches[2:]
		err := serviceConfig.LocationStorage.setLocation(m.Author.ID, strings.Join(l, " "))
		if err != nil {
			fmtstr := fmt.Sprintf("```%s```", fmt.Sprintf("an error occured (%s)", err))
			s.ChannelMessageSend(m.ChannelID, fmtstr)
		}
	}
}
