package main

import (
	"log"
	"net"
	"strconv"
)

var srcIP net.IP = net.IPv4(0, 0, 0, 0)
var srcPort int = 0
var dstIP net.IP = net.IPv4(0, 0, 0, 0)
var dstPort int = 0
var protocoltpye string

func setDefaultConf() {
	srcIP = net.IPv4(0, 0, 0, 0)
	srcPort = 0
	dstIP = net.IPv4(0, 0, 0, 0)
	dstPort = 0
	protocoltpye = ""
	log.Println("Set to default configuration. Protocoltype: " + protocoltpye + ", Source-Port: " + strconv.Itoa(srcPort) + ", Desination-IP: " + dstIP.String() + ", Desination-Port: " + strconv.Itoa(dstPort))
}

func setWar3Conf() {
	srcPort = 6112
	dstPort = 6112
	protocoltpye = "UDP"
	log.Println("Set Warcraft 3 configuration. Protocoltype: " + protocoltpye + ", Source-Port: " + strconv.Itoa(srcPort) + ", Desination-Port: " + strconv.Itoa(dstPort))
}
