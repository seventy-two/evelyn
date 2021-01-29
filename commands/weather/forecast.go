package weather

import (
	"fmt"
	"strings"
	"time"

	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
)

func forecast(location string) (msg string, err error) {
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
	if data.Flags.Units == "us" {
		units = "°F"
	}

	output := fmt.Sprintf("Forecast: %s (%s)\n", location, coords)
	var forecasts []string
	for i := range data.Daily.Data[0:4] {
		tm := time.Unix(data.Daily.Data[i].Time, 0)
		loc, _ := time.LoadLocation(data.Timezone)
		day := tm.In(loc).Weekday()
		forecasts = append(forecasts, fmt.Sprintf("\n%s | %s | %v%s|/|%v%s ",
			day,
			emoji(data.Daily.Data[i].Icon),
			round(data.Daily.Data[i].TemperatureMax),
			units,
			round(data.Daily.Data[i].TemperatureMin),
			units,
		))
	}
	output += columnize.SimpleFormat(forecasts)
	return output, nil
}
