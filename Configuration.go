package main

import (
	"net"
	"strconv"
)

type config struct {
	srcIP       net.IP   // the address to reach default gateway
	pcapDevName string   // pcap name of interface having srcIP to capture from
	winDevName  string   // windows name of interface having srcIP to show in GUI
	gameServer  []net.IP // targeted unicast address, instead broadcast
	gamePorts   []int    // list of game server ports
}

func newDefaultConf() *config {
	config := config{
		nil,
		"",
		"",
		make([]net.IP, 10),
		[]int{
			6112,  // wc3
			28960, // cod1
			27015, // cs1.5
		},
	}
	for i := 0; i < len(config.gameServer); i++ {
		config.gameServer[i] = net.IPv4(10, 0, 0, byte(i+1))
	}
	return &config
}

func (cfg config) gamePorts2StringList() string {
	result := ""
	for _, p := range cfg.gamePorts {
		result += strconv.Itoa(p)
		result += ", "
	}
	result = result[0 : len(result)-2]
	return result
}
