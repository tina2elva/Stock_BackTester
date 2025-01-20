package visualization

import (
	"os"
	"stock/common"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Chart struct {
	title string
}

func NewChart(title string) *Chart {
	return &Chart{title: title}
}

func (c *Chart) PlotCandlestick(data []common.Candle, tradesMap map[string][]common.Trade, outputFile string) error {
	// 创建K线图
	kline := charts.NewKLine()

	// 准备K线数据
	x := make([]string, 0, len(data))
	y := make([]opts.KlineData, 0, len(data))
	for _, candle := range data {
		x = append(x, candle.Timestamp.Format("2006-01-02"))
		y = append(y, opts.KlineData{
			Value: [4]float32{
				float32(candle.Open),
				float32(candle.Close),
				float32(candle.Low),
				float32(candle.High),
			},
		})
	}

	// 准备买卖点数据
	//colors := []string{"green", "blue", "orange", "purple", "brown"}
	legendData := []string{"K线"}
	scatterSeries := make([]*charts.Scatter, 0)

	for strategyName, trades := range tradesMap {
		//color := colors[len(scatterSeries)%len(colors)]
		buyPoints := make([]opts.ScatterData, 0)
		sellPoints := make([]opts.ScatterData, 0)

		for _, trade := range trades {
			date := trade.Timestamp.Format("2006-01-02")
			price := float32(trade.Price)
			if trade.Type == common.ActionBuy {
				buyPoints = append(buyPoints, opts.ScatterData{
					Value:      []interface{}{date, price},
					Symbol:     "circle",
					SymbolSize: 10,
				})
			} else if trade.Type == common.ActionSell {
				sellPoints = append(sellPoints, opts.ScatterData{
					Value:      []interface{}{date, price},
					Symbol:     "circle",
					SymbolSize: 10,
				})
			}
		}

		// 创建散点图用于买卖点
		scatter := charts.NewScatter()
		scatter.SetXAxis(x).
			AddSeries(strategyName+" 买入", buyPoints).
			AddSeries(strategyName+" 卖出", sellPoints)

		scatterSeries = append(scatterSeries, scatter)
		legendData = append(legendData, strategyName+" 买入", strategyName+" 卖出")
	}

	// 设置K线图选项
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: c.title,
			Left:  "center",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "日期",
			Type: "category",
			AxisLabel: &opts.AxisLabel{
				Rotate: 45,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "价格",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
	)

	// 添加K线数据
	kline.SetXAxis(x).AddSeries("K线", y).SetSeriesOptions(
		charts.WithItemStyleOpts(opts.ItemStyle{
			Color:        "#ec0000",
			Color0:       "#00da3c",
			BorderColor:  "#8A0000",
			BorderColor0: "#008F28",
		}),
	)

	// 组合图表
	page := components.NewPage()
	chartsToAdd := []components.Charter{kline}
	for _, scatter := range scatterSeries {
		chartsToAdd = append(chartsToAdd, scatter)
	}
	page.AddCharts(chartsToAdd...)

	// 保存图表
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// 使用HTML渲染器
	return page.Render(f)
}
