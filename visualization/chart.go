package visualization

import (
	"image/color"
	"stock/backtest"
	"stock/common"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type Chart struct {
	title  string
	width  vg.Length
	height vg.Length
}

func NewChart(title string) *Chart {
	return &Chart{
		title:  title,
		width:  20 * vg.Centimeter,
		height: 10 * vg.Centimeter,
	}
}

func (c *Chart) PlotEquityCurve(results *backtest.BacktestResults, filePath string) error {
	p := plot.New()
	p.Title.Text = c.title
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Equity"

	pts := make(plotter.XYs, len(results.EquityCurve))
	for i, v := range results.EquityCurve {
		pts[i].X = float64(i)
		pts[i].Y = v
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}

	p.Add(line)
	p.Legend.Add("Equity", line)

	return p.Save(c.width, c.height, filePath)
}

func (c *Chart) PlotCandlestick(data []*common.DataPoint, filePath string) error {
	p := plot.New()
	p.Title.Text = c.title
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Price"

	// Create candlestick plot using lines
	for i, d := range data {
		// Determine color based on price movement
		var lineColor color.Color
		if d.Close >= d.Open {
			lineColor = color.RGBA{R: 0, G: 128, B: 0, A: 255} // Green for up
		} else {
			lineColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red for down
		}

		// Draw high-low line
		highLow := plotter.XYs{
			{X: float64(i), Y: d.Low},
			{X: float64(i), Y: d.High},
		}
		hlLine, err := plotter.NewLine(highLow)
		if err != nil {
			return err
		}
		hlLine.LineStyle.Width = vg.Points(1)
		hlLine.LineStyle.Color = lineColor
		p.Add(hlLine)

		// Draw open-close box
		openClose := plotter.XYs{
			{X: float64(i) - 0.2, Y: d.Open},
			{X: float64(i) + 0.2, Y: d.Open},
			{X: float64(i) + 0.2, Y: d.Close},
			{X: float64(i) - 0.2, Y: d.Close},
			{X: float64(i) - 0.2, Y: d.Open},
		}
		ocLine, err := plotter.NewLine(openClose)
		if err != nil {
			return err
		}
		ocLine.LineStyle.Width = vg.Points(2)
		ocLine.LineStyle.Color = lineColor
		p.Add(ocLine)
	}

	return p.Save(c.width, c.height, filePath)
}
