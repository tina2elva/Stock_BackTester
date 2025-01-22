module stock

go 1.20

require (
	github.com/go-echarts/go-echarts/v2 v2.2.4
	stock/common/types v0.0.0-00010101000000-000000000000
)

replace stock/common/types => ./common/types/
