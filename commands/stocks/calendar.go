package stocks

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/ryanuber/columnize"
	"github.com/seventy-two/Cara/web"
)

var relevantTickers = []string{"SCHW", "FBK", "BAC", "UAL", "LMT", "IBKR",
	"GS", "JNJ", "ERIC", "TSLA", "PG", "NFLX", "MS",
	"LRCX", "LVS", "USB", "NDAQ", "ZION", "STLD", "T",
	"AAL", "TSM", "NOK", "WDFC", "FCX", "UNP",
	"AXP", "HBAN", "IPG"}

func GetCalendar() string {
	data, err := web.GetBody(fmt.Sprintf(serviceConfig.EarningsURL, serviceConfig.APIKey))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	e, err := translateCSV(data)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var newE []Earnings
	for _, t := range e {
		for _, ticker := range relevantTickers {
			if t.Symbol == ticker {
				newE = append(newE, t)
			}
		}
	}

	sort.Slice(newE, func(i, j int) bool {
		return newE[i].ReportDate.Before(newE[j].ReportDate)
	})

	var lines []string

	for _, l := range newE {
		lines = append(lines, fmt.Sprintf("%s|%s|%s", l.Symbol, l.Name, l.ReportDate.Format("Monday")))
	}
	return fmt.Sprintf("```%s```", columnize.SimpleFormat(lines))
}

func translateCSV(data []byte) ([]Earnings, error) {
	var earnings []Earnings
	r := csv.NewReader(bytes.NewReader(data))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if record[0] != record[1] {
			reportDate, _ := time.Parse("2006-01-02", record[2])
			fiscalDateEnding, _ := time.Parse("2006-01-02", record[3])
			estimate, _ := strconv.ParseFloat(record[4], 32)
			earnings = append(earnings, Earnings{
				Symbol:           record[0],
				Name:             record[1],
				ReportDate:       reportDate,
				FiscalDateEnding: fiscalDateEnding,
				Estimate:         float32(estimate),
				Currency:         record[5],
			})
		}
	}
	return earnings, nil
}
