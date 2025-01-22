package visualization

import (
	"os"
	"stock/common/types"

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

func (c *Chart) PlotCandlestick(data []types.Candle, tradesMap map[string][]types.Trade, outputFile string) error {
	// 创建页面
	page := components.NewPage()

	// 创建K线图
	kline := charts.NewKLine()

	// 创建交易量柱状图
	volume := charts.NewBar()
	volume.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "交易量",
			Left:  "center",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "日期",
			Type: "category",
			AxisLabel: &opts.AxisLabel{
				Rotate: 45,
			},
			GridIndex: 0,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:      "交易量",
			GridIndex: 0,
		}),
	)

	// 创建MACD图表
	macdChart := charts.NewLine()
	macdChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "MACD",
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
			Name: "MACD",
		}),
	)

	// 创建RSI图表
	rsiChart := charts.NewLine()
	rsiChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "RSI",
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
			Name: "RSI",
		}),
	)

	// 准备K线数据
	x := make([]string, 0, len(data))
	y := make([]opts.KlineData, 0, len(data))
	volumeData := make([]opts.BarData, 0, len(data))
	macdLineData := make([]opts.LineData, 0, len(data))
	signalLineData := make([]opts.LineData, 0, len(data))
	histogramData := make([]opts.BarData, 0, len(data))
	rsiData := make([]opts.LineData, 0, len(data))

	for _, candle := range data {
		date := candle.Timestamp.Format("2006-01-02")
		x = append(x, date)
		y = append(y, opts.KlineData{
			Value: [4]float32{
				float32(candle.Open),
				float32(candle.Close),
				float32(candle.Low),
				float32(candle.High),
			},
		})
		// 根据涨跌设置交易量颜色
		if candle.Close > candle.Open {
			// 上涨 - 绿色
			volumeData = append(volumeData, opts.BarData{
				Value: float32(candle.Volume),
				ItemStyle: &opts.ItemStyle{
					Color: "#00da3c",
				},
			})
		} else {
			// 下跌 - 红色
			volumeData = append(volumeData, opts.BarData{
				Value: float32(candle.Volume),
				ItemStyle: &opts.ItemStyle{
					Color: "#ec0000",
				},
			})
		}
		if macdValues, ok := candle.Indicators["MACD"]; ok {
			if macdMap, ok := macdValues.(map[string][]float64); ok {
				for name, values := range macdMap {
					if len(values) > 0 {
						switch name {
						case "MACD":
							macdLineData = append(macdLineData, opts.LineData{
								Value: float32(values[len(values)-1]),
								Name:  "MACD线",
							})
						case "Signal":
							signalLineData = append(signalLineData, opts.LineData{
								Value: float32(values[len(values)-1]),
								Name:  "信号线",
							})
						case "Histogram":
							histogramData = append(histogramData, opts.BarData{
								Value: float32(values[len(values)-1]),
							})
						}
					}
				}
			}
		}
		if rsiValues, ok := candle.Indicators["RSI"]; ok {
			if rsiMap, ok := rsiValues.(map[string][]float64); ok {
				for name, values := range rsiMap {
					if len(values) > 0 {
						rsiData = append(rsiData, opts.LineData{
							Value: float32(values[len(values)-1]),
							Name:  name,
						})
					}
				}
			}
		}
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
			if trade.Type == types.ActionBuy {
				buyPoints = append(buyPoints, opts.ScatterData{
					Value:      []interface{}{date, price},
					Symbol:     "circle",
					SymbolSize: 10,
				})
			} else if trade.Type == types.ActionSell {
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
			GridIndex: 0,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:      "价格",
			GridIndex: 0,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0, 1},
			YAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0, 1},
			YAxisIndex: []int{0},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "800px",
			Theme:  "light",
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

	// 设置图表数据
	kline.SetXAxis(x).AddSeries("K线", y).SetSeriesOptions(
		charts.WithItemStyleOpts(opts.ItemStyle{
			Color:        "#ec0000",
			Color0:       "#00da3c",
			BorderColor:  "#8A0000",
			BorderColor0: "#008F28",
		}),
	)

	volume.SetXAxis(x).AddSeries("交易量", volumeData)
	// 创建MACD柱状图
	macdBar := charts.NewBar()
	macdBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "MACD柱状图",
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
			Name: "MACD柱状图",
		}),
	)
	macdBar.SetXAxis(x).AddSeries("MACD柱状图", histogramData)

	macdChart.SetXAxis(x).
		AddSeries("MACD线", macdLineData).
		AddSeries("信号线", signalLineData)
	rsiChart.SetXAxis(x).AddSeries("RSI", rsiData)

	// 组合图表
	chartsToAdd := []components.Charter{kline, volume, macdBar, macdChart, rsiChart}
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
