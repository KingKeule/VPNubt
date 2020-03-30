package main

import "net"

func main() {

	initService()
	//InitGUI()

	// TODO from gui
	currentConfig.pcapPortUDP = 6112
	currentConfig.VPNHostTarget = net.IP{192, 168, 1, 2}

	startServicePcap() // TODO from gui, new thread
}
