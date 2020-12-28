package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/tatsushid/go-fastping"
)

var currentConfig *config = nil

type helperIPv4Routing struct {
	target  net.IP
	mask    net.IP
	gateway net.IP
	iface   net.IP
	metric  int
}

func getWin32Routes() ([]helperIPv4Routing, error) {

	routes := make([]helperIPv4Routing, 0)

	cmd := exec.Command("route", "print")
	r, w := io.Pipe()
	cmd.Stdout = w

	go func() {
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	scanner := bufio.NewScanner(r)
	routeLinesFound := false
	routeLinesOffset := -4 // distance from line matched "IPv4-Routentabelle" to first line with table data
	routeLineSplit := regexp.MustCompile(`\s{2,}`)
	for scanner.Scan() {
		ucl := scanner.Text()

		if 0 == routeLinesOffset {

			if strings.Contains(ucl, "=") {
				break
			}

			split := routeLineSplit.Split(ucl, -1)
			//fmt.Printf("%#v\n", split)

			metric, _ := strconv.Atoi(split[5])

			routes = append(routes,
				helperIPv4Routing{
					target:  net.ParseIP(split[1]),
					mask:    net.ParseIP(split[2]),
					gateway: net.ParseIP(split[3]),
					iface:   net.ParseIP(split[4]),
					metric:  metric})

		} else if !routeLinesFound {
			if strings.Contains(ucl, "IPv4-Routentabelle") {
				routeLinesFound = true
				routeLinesOffset++
			}
		} else {
			routeLinesOffset++
		}
	}

	//fmt.Printf("%s", routes[0])

	if err := scanner.Err(); err != nil {
		//fmt.Fprintln(os.Stderr, "error:", err)
		return routes, err
	}

	return routes, nil
}

// checks whether the given ip host address can be pinged
func ping(addr string) (bool, error) {

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", addr /*inputIP.Text*/)
	var recieved bool = false

	if err != nil {
		log.Println(err)
		return recieved, err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		recieved = true
		log.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}

	p.OnIdle = func() {
		log.Println("Ping job finished")
	}

	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	return recieved, nil
}

func anyPcapAddress(these []pcap.InterfaceAddress, other net.IP) bool {
	for _, cur := range these {
		if cur.IP.Equal(other) {
			return true
		}
	}
	return false
}

func anyWinNetworkAddress(these []net.Addr, other net.IP) bool {
	for _, cur := range these {
		addrNoPort, _, _ := net.ParseCIDR(cur.String())
		if addrNoPort.Equal(other) {
			return true
		}
	}
	return false
}

func autoSelectNetworkInterface() error {
	routes, errRoutes := getWin32Routes()
	if errRoutes != nil {
		return errRoutes
	}
	if len(routes) == 0 {
		return fmt.Errorf("Empty routing table parsed")
	}

	// routes already sorted by metric, so take first
	autoIfaceIP := routes[0].iface
	currentConfig.srcIP = autoIfaceIP

	log.Printf("Using IP '%s' to find pcap device\n", currentConfig.srcIP.String())

	allPcapDevs, errPcapDevs := pcap.FindAllDevs()
	if errPcapDevs != nil {
		log.Fatal(errPcapDevs)
		return errPcapDevs
	}
	for _, pcapDev := range allPcapDevs {
		if anyPcapAddress(pcapDev.Addresses, autoIfaceIP) {
			currentConfig.pcapDevName = pcapDev.Name
			break
		}
	}
	if len(currentConfig.pcapDevName) == 0 {
		return fmt.Errorf("No suitable pcap device found")
	}

	allWinNetworkDevs, errWinNetworkDecs := net.Interfaces()
	if errWinNetworkDecs != nil {
		log.Fatal(errWinNetworkDecs)
		return errWinNetworkDecs
	}
	for _, winNetworkDev := range allWinNetworkDevs {
		winAddr, err := winNetworkDev.Addrs()
		if err != nil {
			log.Printf("Skipping Device '%s' due to Error getting Addresses\n", winNetworkDev.Name)
			continue
		}
		if anyWinNetworkAddress(winAddr, autoIfaceIP) {
			currentConfig.winDevName = winNetworkDev.Name
			break
		}
	}
	if len(currentConfig.winDevName) == 0 {
		return fmt.Errorf("No suitable win network device found")
	}

	return nil
}

// capute all udp broadcast packets on given port and network device
// via the StopThreadChannel this function receives the information from the GUI to be stopped
func capturePackets(stopThreadChannel chan bool) {

	const snapshotLength int32 = 1024 // the maximum size to read for each packet
	const promiscuous bool = false    // interface in promiscuous mode

	// create a pcap handle for given network device
	handle, err := pcap.OpenLive(currentConfig.pcapDevName, snapshotLength, promiscuous, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}

	// defers the pcap execution until the surrounding function (capture + forward) returns
	defer handle.Close()

	// Set pcap filter
	var filterBuild strings.Builder

	fmt.Fprintf(&filterBuild, "src host %s", currentConfig.srcIP)
	fmt.Fprint(&filterBuild, " and ip broadcast")
	fmt.Fprint(&filterBuild, " and (udp port")
	for i, gp := range currentConfig.gamePorts {
		fmt.Fprintf(&filterBuild, " %d", gp)
		if i != len(currentConfig.gamePorts)-1 {
			fmt.Fprint(&filterBuild, " or")
		}
	}
	fmt.Fprint(&filterBuild, ")")

	filter := filterBuild.String()
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set pcap filter to: '" + filter + "'")

	// Use the handle as a packet source to process all packets
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	//selection, whether to start or stop the service when the stop signal comes
	for {
		select {
		case packet := <-packets:
			// forward each captured packet
			log.Println("UDP broadcast packet was captured and will be forwarded as udp unicast")
			forwardPacket(packet)
		case <-stopThreadChannel:
			log.Println("Stop signal recieved")
			return
		}
	}
}

// send the captured broacast packet as unicast to the given ip adress
func forwardPacket(packet gopacket.Packet) {

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		log.Fatalf("Unable to get UDP Layer of a Packet captured with UDP Filter?!")
		return
	}
	udp, _ := udpLayer.(*layers.UDP)
	dstPort := udp.DstPort

	for _, gameServerIP := range currentConfig.gameServer {

		expr := fmt.Sprintf("%s:%d", gameServerIP.String(), dstPort)
		serverAddr, err := net.ResolveUDPAddr("udp4", expr)
		if err != nil {
			log.Fatal(err)
		}

		expr = fmt.Sprintf(":%d", dstPort)
		conn, err := net.ListenPacket("udp", expr)
		if err != nil {
			log.Fatal(err)
		}

		// defers the udp forward execution until the surrounding function (udp connection) returns
		defer conn.Close()

		// get payload from captured packet
		data := packet.ApplicationLayer().Payload()

		// send new unicast packet to server
		_, err = conn.WriteTo(data, serverAddr)
		if err != nil {
			log.Println("Packet was not successfully forwarded as udp unicast to: " + gameServerIP.String())
			log.Fatal(err)
		} else {
			log.Println("Packet was successfully forwarded as udp unicast to: " + gameServerIP.String())
		}
	}
}
