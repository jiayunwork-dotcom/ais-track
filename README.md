# ais-track — 船舶 AIS 航迹分析 HTTP 服务与 CLI

解析 AIS 位置报、检测超速/转向/报告间隔/港域滞留，并计算两船最近会遇点（CPA）。
提供 CLI 批处理与 HTTP API。

## 构建与运行

```sh
go build -o ais-track .
./ais-track serve -addr :8080
./ais-track -input example/sample.csv -maxsog 30
go test ./...
```

未提供 `-input` 且未通过管道传入数据时，CLI 以用法错误（退出码 2）结束。

## HTTP

- `GET /api/health` — `{"status":"ok"}`
- `POST /api/analyze` — JSON 航迹点，返回超速计数与异常列表；非法坐标 HTTP 422
- `POST /api/cpa` — 两船航迹，返回 DCPA（km）与会遇时刻；无时间重叠 HTTP 422

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
- **转向 / 报告间隔**：航向突变与超阈时间间隔分开处理；跨大间隔不做航向插值。

## 许可

MIT — 见 [LICENSE](./LICENSE)。
