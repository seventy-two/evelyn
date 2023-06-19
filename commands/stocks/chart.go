package stocks

import (
	"bytes"
	"fmt"
	"image/color"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pplcc/plotext/custplotter"
	"github.com/seventy-two/Cara/web"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func makeChart(s *discordgo.Session, m *discordgo.MessageCreate, q string) {
	d := &ChartData{}
	url := fmt.Sprintf(serviceConfig.ChartURL, q, serviceConfig.APIKey)
	err := web.GetJSON(url, d)
	if err != nil {
		fmt.Println(err)
		return
	}
	data := createData(*d)

	p := plot.New()
	p.BackgroundColor = color.RGBA{R: 155, G: 155, B: 155, A: 255}
	p.X.Padding = vg.Length(5) * vg.Points(1)
	p.Y.Padding = vg.Length(5) * vg.Points(1)
	p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

	bars, err := custplotter.NewCandlesticks(data)
	if err != nil {
		fmt.Println(err)
		//log.Panic(err)
	}
	bars.ColorUp = color.RGBA{R: 140, G: 193, B: 118, A: 255}
	bars.ColorDown = color.RGBA{R: 184, G: 44, B: 12, A: 255}
	bars.CandleWidth = vg.Length(5) * plotter.DefaultLineStyle.Width

	p.Add(bars)

	c := vgimg.PngCanvas{
		Canvas: vgimg.NewWith(
			vgimg.UseWH(1200, 700),
			vgimg.UseBackgroundColor(color.RGBA64{R: 84, G: 87, B: 93, A: 255}),
		)}
	p.Draw(draw.New(c))

	//w, err := p.WriterTo(1200, 700, "png")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	var b []byte
	buf := bytes.NewBuffer(b)
	c.WriteTo(buf)
	s.ChannelFileSend(m.ChannelID, "chart.png", buf)
}

func createData(d ChartData) custplotter.TOHLCVs {
	data := make(custplotter.TOHLCVs, len(d))
	for i := range d {
		t, err := time.Parse("15:04", d[i].Minute)
		if err != nil {
			continue
		}
		data[i].T = float64(t.Unix())
		if d[i].Volume == 0 && i != 0 {
			data[i].O = data[i-1].C
			data[i].C = data[i-1].C
			data[i].H = data[i-1].C
			data[i].L = data[i-1].C
			continue
		}
		data[i].O = d[i].Open
		data[i].C = d[i].Close
		data[i].H = d[i].High
		data[i].L = d[i].Low
	}
	return data
}
