package datasource

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"stock/common/types"
)

// PeriodType 周期类型
type PeriodType int

const (
	PeriodTypeMinute PeriodType = iota
	PeriodTypeHour
	PeriodTypeDay
	PeriodTypeWeek
	PeriodTypeMonth
)

// DataSource 数据源接口
type DataSource interface {
	// 获取指定周期的数据
	GetData(symbol string, period PeriodType, start, end time.Time) ([]*types.DataPoint, error)

	// 获取支持的所有周期
	GetSupportedPeriods() []PeriodType

	// 将数据转换为指定周期
	ConvertPeriod(data []*types.DataPoint, targetPeriod PeriodType) ([]*types.DataPoint, error)
}

// CSVDataSource CSV文件数据源
type CSVDataSource struct {
	path string
}

func NewCSVDataSource(path string) *CSVDataSource {
	return &CSVDataSource{path: path}
}

// DetectPeriod 根据数据时间间隔自动判断周期
func (ds *CSVDataSource) DetectPeriod(data []*types.DataPoint) PeriodType {
	if len(data) < 2 {
		return PeriodTypeDay
	}

	// 计算时间间隔
	interval := data[1].Timestamp.Sub(data[0].Timestamp)

	switch {
	case interval < time.Hour:
		return PeriodTypeMinute
	case interval < 24*time.Hour:
		return PeriodTypeHour
	case interval < 7*24*time.Hour:
		return PeriodTypeDay
	case interval < 30*24*time.Hour:
		return PeriodTypeWeek
	default:
		return PeriodTypeMonth
	}
}

func (ds *CSVDataSource) GetData(symbol string, period PeriodType, start, end time.Time) ([]*types.DataPoint, error) {
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

	var points []*types.DataPoint
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

			points = append(points, &types.DataPoint{
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

func (ds *CSVDataSource) GetSupportedPeriods() []PeriodType {
	return []PeriodType{PeriodTypeDay}
}

func (ds *CSVDataSource) ConvertPeriod(data []*types.DataPoint, targetPeriod PeriodType) ([]*types.DataPoint, error) {
	if targetPeriod == PeriodTypeDay {
		return data, nil
	}

	// 按周转换
	if targetPeriod == PeriodTypeWeek {
		return convertToWeekly(data)
	}

	return nil, fmt.Errorf("unsupported period conversion: %v", targetPeriod)
}

func convertToWeekly(data []*types.DataPoint) ([]*types.DataPoint, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var weeklyData []*types.DataPoint
	var currentWeek *types.DataPoint

	for _, dp := range data {
		year, week := dp.Timestamp.ISOWeek()
		currentYear, currentWeekNum := currentWeek.Timestamp.ISOWeek()
		if currentWeek == nil || currentYear != year || currentWeekNum != week {
			if currentWeek != nil {
				weeklyData = append(weeklyData, currentWeek)
			}
			currentWeek = &types.DataPoint{
				Timestamp:  dp.Timestamp,
				Open:       dp.Open,
				High:       dp.High,
				Low:        dp.Low,
				Close:      dp.Close,
				Volume:     dp.Volume,
				Indicators: make(map[string]float64),
			}
		} else {
			currentWeek.High = math.Max(currentWeek.High, dp.High)
			currentWeek.Low = math.Min(currentWeek.Low, dp.Low)
			currentWeek.Close = dp.Close
			currentWeek.Volume += dp.Volume
		}
	}

	if currentWeek != nil {
		weeklyData = append(weeklyData, currentWeek)
	}

	return weeklyData, nil
}
