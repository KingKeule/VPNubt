package main

import (
	"net"
)

type config struct {
	srcIP   net.IP
	dstIP   net.IP
	srcPort int
	dstPort int
}

func getDefaultConf() *config {
	config := config{net.IPv4(0, 0, 0, 0), net.IPv4(0, 0, 0, 0), 0, 0}
	return &config
}

func getWar3Conf() *config {
	config := config{nil, nil, 6112, 6112}
	return &config
}

func getCoDUOConf() *config {
	config := config{nil, nil, 28960, 28960}
	return &config
}
