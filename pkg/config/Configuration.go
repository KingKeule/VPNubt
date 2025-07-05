package config

import (
	"net/netip"

	"fyne.io/fyne"
)

var Prefs fyne.Preferences

func InitAfterFyneApp() {
	network := Prefs.StringWithFallback("network", "new")
	if network == "new" {
		Prefs.SetString("network", "10.0.0.0/29")
	}
	thisPeerIP := Prefs.StringWithFallback("thisPeerIP", "new")
	if thisPeerIP == "new" {
		Prefs.SetString("thisPeerIP", "10.0.0.1")
	}
}

func ThisNetworkHostAddresses() []string {
	result := make([]string, 0, 10)
	network := Prefs.String("network")
	hosts, _ := hosts(network)
	for _, h := range hosts {
		result = append(result, h.String())
	}
	return result
}

func hosts(cidr string) ([]netip.Addr, error) {
	var ips []netip.Addr
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return ips, err
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		ips = append(ips, addr)
	}

	if len(ips) < 2 {
		return ips, nil
	}

	return ips[1 : len(ips)-1], nil
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
