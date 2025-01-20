package datasource

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"

	"stock/common"
)

// DataSource 数据源接口
type DataSource interface {
	GetData(symbol string, start, end time.Time) ([]*common.DataPoint, error)
}

// CSVDataSource CSV文件数据源
type CSVDataSource struct {
	path string
}

func NewCSVDataSource(path string) *CSVDataSource {
	return &CSVDataSource{path: path}
}

func (ds *CSVDataSource) GetData(symbol string, start, end time.Time) ([]*common.DataPoint, error) {
	file, err := os.Open(ds.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var points []*common.DataPoint
	var closePrices []float64

	for _, record := range records[1:] { // 跳过表头
		timestamp, _ := time.Parse("2006-01-02", record[0])
		open, _ := strconv.ParseFloat(record[1], 64)
		close, _ := strconv.ParseFloat(record[2], 64)
		high, _ := strconv.ParseFloat(record[3], 64)
		low, _ := strconv.ParseFloat(record[4], 64)
		volume, _ := strconv.ParseFloat(record[5], 64)

		if timestamp.After(start) && timestamp.Before(end) {
			closePrices = append(closePrices, close)
			ma5 := calculateMA(closePrices, 5)

			points = append(points, &common.DataPoint{
				Timestamp: timestamp,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
				Indicators: map[string]float64{
					"MA5": ma5,
				},
			})
		}
	}

	return points, nil
}

// calculateMA 计算移动平均线
func calculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	var sum float64
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}
