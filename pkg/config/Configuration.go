package config

import "fyne.io/fyne"

var Prefs fyne.Preferences

func InitAfterFyneApp() {
	network := Prefs.StringWithFallback("network", "new")
	if network == "new" {
		Prefs.SetString("network", "10.0.0.0/29")
	}
}

type Config struct {
	DstIP   string `json:"IP"`
	DstPort int    `json:"Port"`
}

func GetDefaultConf() *Config {
	config := Config{"0.0.0.0", 0}
	return &config
}

func GetWar3Conf() *Config {
	config := Config{"", 6112}
	return &config
}

func GetCoDUOConf() *Config {
	config := Config{"", 28960}
	return &config
}
