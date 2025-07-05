package service

import (
	"github.com/KingKeule/VPNubt/pkg/config"
	"golang.zx2c4.com/wireguard/windows/conf"
)

func newWgConf() string {
	pk, _ := conf.NewPrivateKey()
	config := conf.Config{
		Name:      "VPNubt",
		Interface: conf.Interface{PrivateKey: *pk},
	}
	return config.ToWgQuick()
}

func WgConf() *conf.Config {
	asString := config.Prefs.StringWithFallback("wgconf", "new")
	if asString == "new" {
		asString = newWgConf()
	}
	wgconf, _ := conf.FromWgQuick(asString, "VPNubt")
	return wgconf
}
