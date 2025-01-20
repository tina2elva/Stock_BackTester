import requests
import pandas as pd
import time
from retrying import retry

@retry(stop_max_attempt_number=3, wait_fixed=2000)
def fetch_data():
    # 东方财富API参数
    stock_code = "600036"  # 招商银行A股
    start_date = "20100101"
    end_date = time.strftime("%Y%m%d", time.localtime())

    # 获取K线数据
    url = f"http://push2his.eastmoney.com/api/qt/stock/kline/get"
    params = {
        "secid": f"1.{stock_code}",  # 1表示沪市
        "fields1": "f1,f2,f3,f4,f5",
        "fields2": "f51,f52,f53,f54,f55,f56,f57",
        "klt": "101",  # 日K线
        "fqt": "1",  # 前复权
        "beg": start_date,
        "end": end_date,
        "lmt": "10000"  # 最大数据量
    }

    response = requests.get(url, params=params, timeout=10)
    response.raise_for_status()
    return response.json()

try:
    # 获取数据
    data = fetch_data()["data"]["klines"]

    # 解析数据
    parsed_data = []
    for item in data:
        parts = item.split(",")
        parsed_data.append({
            "Date": parts[0],
            "Open": float(parts[1]),
            "Close": float(parts[2]),
            "High": float(parts[3]),
            "Low": float(parts[4]),
            "Volume": int(float(parts[5])),
            "Amount": float(parts[6])  # 成交额
        })

    # 转换为DataFrame
    df = pd.DataFrame(parsed_data)

    # 保存为CSV文件
    df.to_csv('data/cmb.csv', index=False)
    print("数据获取成功，已保存到 data/cmb.csv")
    
except Exception as e:
    print(f"数据获取失败: {str(e)}")
