# ais-track: Go headless HTTP API and CLI for AIS vessel track analysis

用户给出 AIS CSV/JSON 航迹点（MMSI、时间、经纬度、SOG、COG）、超速阈值与可选港域多边形，工具算出超速计数、转向突变、报告间隔与港内滞留异常，并对两船航迹插值后给出最近会遇距离 DCPA 与会遇时刻 TCPA。坐标越界必须报错；报告间隔超过阈值时不做航向插值；港域内连续 ≥3 条记为滞留；空或截断的 JSON 归档读回必须拒绝。

## How to run

```
go run . serve -addr :8080
go run . -input example/sample.csv -maxsog 30
```

## API

- `GET /api/health` — liveness `{"status":"ok"}`.
- `POST /api/analyze` — records and max_sog return speeding count and anomalies. Illegal lat/lon → 422.
- `POST /api/cpa` — own and target tracks return DCPA (km) and the CPA timestamp. No overlap → 422.

## Build

```
go build ./...
go test ./...
```

Module targets Go 1.21.

## Evaluation Image

Evaluation-specific files (do not overwrite project Dockerfile/README):

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md` (this file)

Build:

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
```
