# armbian-stats

Lightweight system monitor for Armbian / Linux SBCs.
Single static binary, no external dependencies at runtime.

## Build

```bash
go mod tidy
go build -o armbian-stats .
./armbian-stats
```

Open http://localhost:8080

### Cross-compile

```bash
# arm64 (OrangePi, RockPi, etc.)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o armbian-stats .

# armv7
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o armbian-stats .
```

## Configuration

`config.yml` is created automatically on first run.

```yaml
host: "0.0.0.0"
port: 8080
interval: 2

theme:
  background:  "#0d1117"
  surface:     "#161b22"
  surface_alt: "#1c2128"
  primary:     "#58a6ff"
  secondary:   "#3fb950"
  accent:      "#f0883e"
  warning:     "#ff7b72"
  text:        "#e6edf3"
  text_muted:  "#8b949e"
  border:      "#30363d"
```

### Nord theme

```yaml
theme:
  background: "#2e3440"
  surface:    "#3b4252"
  primary:    "#88c0d0"
  secondary:  "#a3be8c"
  accent:     "#ebcb8b"
  warning:    "#bf616a"
  text:       "#eceff4"
  text_muted: "#d8dee9"
  border:     "#4c566a"
```

### Dracula theme

```yaml
theme:
  background: "#282a36"
  surface:    "#44475a"
  primary:    "#bd93f9"
  secondary:  "#50fa7b"
  accent:     "#ffb86c"
  warning:    "#ff5555"
  text:       "#f8f8f2"
  text_muted: "#6272a4"
  border:     "#44475a"
```

## API

| Endpoint | Description |
|---|---|
| `GET /` | Web UI |
| `GET /api/stream` | SSE real-time JSON stream |
| `GET /api/stats` | Single JSON snapshot |
| `GET /health` | Health check |

## Data sources

| Panel | Source |
|---|---|
| Temperature | `/sys/class/thermal/thermal_zone*` + `/sys/class/hwmon/hwmon*/temp*_input` |
| CPU % | `/proc/stat` (delta between samples) |
| CPU MHz | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq` |
| RAM / Swap | `/proc/meminfo` |
| Disk | `syscall.Statfs` |
| Network | `/proc/net/dev` |

## systemd

```ini
[Unit]
Description=armbian-stats
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/armbian-stats
ExecStart=/opt/armbian-stats/armbian-stats
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
