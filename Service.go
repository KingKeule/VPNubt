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
	"github.com/google/gopacket/pcap"
	"github.com/tatsushid/go-fastping"
)

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

	err := cmd.Run()
	if err != nil {
		return routes, fmt.Errorf("The command 'route print' went wrong")
	}

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

// get all network interfaces as string list
func getNetworkInterfaces() []string {

	log.Println("Search for available network interfaces")

	// get the names from all network interfaces
	netDeviceList, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	// copy network list to a string array
	var netDeviceListS []string
	for _, iface := range netDeviceList {
		if !strings.Contains(iface.Name, "Loopback") {
			netDeviceListS = append(netDeviceListS, iface.Name)
		}
	}

	log.Printf("Found %d network interfaces", cap(netDeviceListS))

	return netDeviceListS
}

//check if an ip adress from a net interface adress match to a pcap ip adress
func comparePcapAddress(pcapAdresses []pcap.InterfaceAddress, netIPAdress net.IP) bool {

	if nil == netIPAdress || nil == pcapAdresses {
		return false
	}

	// iterates all pacp adresses from a given pcap interface and compare it to a given ip adress
	for _, pcapAdress := range pcapAdresses {
		if pcapAdress.IP.Equal(netIPAdress) {
			return true
		}
	}

	return false
}

// check if the given pcap interface and net interface have the same ip adress
func sameIP(netIface net.Interface, pcapIface pcap.Interface) bool {

	// get all adresses from net interface
	addrs, err := netIface.Addrs()
	if err != nil {
		log.Fatal(err)
		return false
	}

	// iterates all adress from a given net interface
	for _, addr := range addrs {
		addrNoPort, _, _ := net.ParseCIDR(addr.String())
		if comparePcapAddress(pcapIface.Addresses, addrNoPort) {
			return true
		}
	}
	return false
}

// this function returns the required windows network device name because the net network interface name does not work for pcap capture
func getWindowsNetworkDeviceAddr(networkName string) string {

	log.Println("Start searching for the Windows device name (required by pcap) for the selected GUI network interface [" + networkName + "] by IP address")

	result := ""

	// get all net interface
	netIfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	// get the interface object for the selected network name
	var guiSelectedInterface net.Interface
	for _, iface := range netIfaces {
		if iface.Name == networkName {
			guiSelectedInterface = iface
			break
		}
	}

	// get all pcap interfaces
	pcapIfaces, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// iterates all pcap interfaces and check if the ip from pcap interface is the same as from the select gui
	for _, pcapiface := range pcapIfaces {
		if sameIP(guiSelectedInterface, pcapiface) {
			result = pcapiface.Name
			log.Println("Match of an ip address between net interface [" + guiSelectedInterface.Name + "] and pacp interface [" + pcapiface.Description + "]")
			log.Println("The found Windows device name is: " + result)
			break
		} else {
			log.Println("No match of an ip address between net interface [" + guiSelectedInterface.Name + "] and pacp interface [" + pcapiface.Description + "]")
		}
	}

	return result
}

// capute all udp broadcast packets on given port and network device
// via the StopThreadChannel this function receives the information from the GUI to be stopped
func capturePackets(stopThreadChannel chan bool, networkDevice string, dstIP net.IP, dstPort int) {

	const snapshotLength int32 = 1024 // the maximum size to read for each packet
	const promiscuous bool = false    // interface in promiscuous mode

	// create a pcap handle for given network device
	handle, err := pcap.OpenLive(getWindowsNetworkDeviceAddr(networkDevice), snapshotLength, promiscuous, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}

	// defers the pcap execution until the surrounding function (capture + forward) returns
	defer handle.Close()

	// Set pcap filter
	filter := "udp and port " + strconv.Itoa(dstPort) + " and ip broadcast"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set pcap filter to: " + filter + ".")

	// Use the handle as a packet source to process all packets
	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	//selection, whether to start or stop the service when the stop signal comes
	for {
		select {
		case packet := <-packets:
			// forward each captured packet
			log.Println("UDP broadcast packet was captured and will be forwareded as udp unicast to " + dstIP.String())
			forwardPacket(dstIP, dstPort, packet)
		case <-stopThreadChannel:
			log.Println("Stop tunneling signal recieved")
			log.Println("Tunneling service stopped")
			return
		}
	}
}

// send the captured broacast packet as unicast to the given ip adress
func forwardPacket(dstIP net.IP, dstPort int, packet gopacket.Packet) {

	serverAddr, err := net.ResolveUDPAddr("udp4", dstIP.String()+":"+strconv.Itoa(dstPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenPacket("udp", ":"+strconv.Itoa(dstPort))
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
		log.Println("Packet was not successfully forwarded as udp unicast to: " + dstIP.String())
		log.Fatal(err)
	} else {
		log.Println("Packet was successfully forwarded as udp unicast to: " + dstIP.String())
	}
}
