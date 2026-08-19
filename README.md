# ais-track — 船舶 AIS 轨迹分析与异常航行检测

`ais-track` 是一个纯标准库实现的命令行工具，用于解析船舶 AIS 轨迹数据并检测异常航行行为
（超速、港内滞留）。无 Web UI，无外部依赖。

## 构建与运行

```sh
go build -o ais-track .
./ais-track -input example/sample.csv
```

也可以从标准输入读取 CSV：

```sh
cat example/sample.csv | ./ais-track
```

未提供 `-input` 且未通过管道传入数据时，工具以用法错误（退出码 2）结束，不会崩溃。

## CLI 参数

| 参数        | 说明                                            | 默认值 |
| ----------- | ----------------------------------------------- | ------ |
| `-input`    | AIS CSV 输入文件路径（省略则读 stdin）          | 空     |
| `-maxsog`   | 允许的最大对地航速（节），超过记为超速         | 30.0   |
| `-port`     | 可选：港域多边形 CSV 路径（每行 `lat,lon`）     | 空     |
| `-help`     | 显示用法                                        | false  |

## 输入格式（CSV）

首行为表头，字段顺序固定：

```
mmsi,ts,lat,lon,sog,cog
440123456,2023-06-01T08:00:00Z,35.0,129.0,8.5,180.0
```

- `mmsi`：船舶 MMSI 标识
- `ts`：时间戳（RFC3339 或 `2006-01-02 15:04:05` 等常见格式）
- `lat`,`lon`：纬度、经度（十进制度）
- `sog`：对地航速（节）
- `cog`：对地航向（度）

## 检测规则

- **超速 (speeding)**：单条记录 `SOG > maxSOG`，每条产生一个异常。
- **港内滞留 (loitering)**：连续 ≥3 条记录落在 `-port` 多边形内时，产生一个滞留异常。

## 包结构

- `internal/parse` — CSV 解析与按船舶分组
- `internal/geo` — 地理计算（Haversine 距离、点在多边形内）
- `internal/detect` — 异常检测

## 许可

MIT — 见 [LICENSE](./LICENSE)。
