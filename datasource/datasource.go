package datasource

import (
	"encoding/binary"
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
	GetData(symbol string, period PeriodType, start, end time.Time) ([]*types.DataPoint, error)
	GetSupportedPeriods() []PeriodType
	ConvertPeriod(data []*types.DataPoint, targetPeriod PeriodType) ([]*types.DataPoint, error)
}

// CSVDataSource CSV文件数据源
type CSVDataSource struct {
	path string
}

func NewCSVDataSource(path string) *CSVDataSource {
	return &CSVDataSource{path: path}
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

// TDXDataSource 通达信数据源
type TDXDataSource struct {
	path string
}

func NewTDXDataSource(path string) *TDXDataSource {
	return &TDXDataSource{path: path}
}

// TDXDayRecord 通达信日线数据结构
type TDXDayRecord struct {
	Date     uint32
	Open     uint32
	High     uint32
	Low      uint32
	Close    uint32
	Amount   uint32
	Volume   uint32
	Reserved [4]byte
}

func (ds *TDXDataSource) GetData(symbol string, period PeriodType, start, end time.Time) ([]*types.DataPoint, error) {
	file, err := os.Open(ds.path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := fileInfo.Size()

	// 检查文件大小是否合理
	recordSize := int64(binary.Size(TDXDayRecord{}))
	if fileSize%recordSize != 0 {
		return nil, fmt.Errorf("文件大小不匹配，可能已损坏")
	}

	// 读取文件内容
	records := make([]TDXDayRecord, fileSize/recordSize)
	err = binary.Read(file, binary.LittleEndian, &records)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var points []*types.DataPoint
	var closePrices []float64

	for _, record := range records {
		// 转换日期格式
		year := int(record.Date / 10000)
		month := int((record.Date % 10000) / 100)
		day := int(record.Date % 100)
		timestamp := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

		if timestamp.After(start) && timestamp.Before(end) {
			closePrices = append(closePrices, float64(record.Close))
			ma5 := calculateMA(closePrices, 5)

			points = append(points, &types.DataPoint{
				Timestamp: timestamp,
				Open:      float64(record.Open) / 100,
				High:      float64(record.High) / 100,
				Low:       float64(record.Low) / 100,
				Close:     float64(record.Close) / 100,
				Volume:    float64(record.Volume),
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

func (ds *TDXDataSource) GetSupportedPeriods() []PeriodType {
	return []PeriodType{PeriodTypeDay}
}

func (ds *TDXDataSource) ConvertPeriod(data []*types.DataPoint, targetPeriod PeriodType) ([]*types.DataPoint, error) {
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
