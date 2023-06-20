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
	sort.Slice(e, func(i, j int) bool {
		return e[i].FiscalDateEnding.Before(e[j].FiscalDateEnding)
	})

	e = e[1:15]

	var lines []string

	for _, l := range e {
		lines = append(lines, fmt.Sprintf("%s|%s|%s|Fiscal End: %s|Est: %.2f|%s", l.Symbol, l.Name, l.ReportDate.Format("2006-01-02"), l.FiscalDateEnding.Format("2006-01-02"), l.Estimate, l.Currency))
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
