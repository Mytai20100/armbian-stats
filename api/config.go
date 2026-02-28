package api

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Interval int    `yaml:"interval"`
	Theme    Theme  `yaml:"theme"`
}

type Theme struct {
	Background string `yaml:"background"`
	Surface    string `yaml:"surface"`
	SurfaceAlt string `yaml:"surface_alt"`
	Primary    string `yaml:"primary"`
	Secondary  string `yaml:"secondary"`
	Accent     string `yaml:"accent"`
	Warning    string `yaml:"warning"`
	Text       string `yaml:"text"`
	TextMuted  string `yaml:"text_muted"`
	Border     string `yaml:"border"`
}

const defaultConfigYAML = `# armbian-stats configuration

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
`

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("[config] %q not found, creating default\n", path)
		if err := os.WriteFile(path, []byte(defaultConfigYAML), 0644); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.Interval < 1 {
		cfg.Interval = 2
	}
	applyThemeDefaults(&cfg.Theme)

	return &cfg, nil
}

func applyThemeDefaults(t *Theme) {
	d := Theme{
		Background: "#0d1117",
		Surface:    "#161b22",
		SurfaceAlt: "#1c2128",
		Primary:    "#58a6ff",
		Secondary:  "#3fb950",
		Accent:     "#f0883e",
		Warning:    "#ff7b72",
		Text:       "#e6edf3",
		TextMuted:  "#8b949e",
		Border:     "#30363d",
	}
	if t.Background == "" { t.Background = d.Background }
	if t.Surface == "" { t.Surface = d.Surface }
	if t.SurfaceAlt == "" { t.SurfaceAlt = d.SurfaceAlt }
	if t.Primary == "" { t.Primary = d.Primary }
	if t.Secondary == "" { t.Secondary = d.Secondary }
	if t.Accent == "" { t.Accent = d.Accent }
	if t.Warning == "" { t.Warning = d.Warning }
	if t.Text == "" { t.Text = d.Text }
	if t.TextMuted == "" { t.TextMuted = d.TextMuted }
	if t.Border == "" { t.Border = d.Border }
}
