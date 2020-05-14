package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/tatsushid/go-fastping"
)

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
		log.Println(err)
	}

	return recieved, nil
}

func getNetworkDevices() []string {

	// get the names from all network interfaces
	netDeviceList, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	// copy network list to a string array
	var netDeviceListS []string
	for _, param := range netDeviceList {
		if !strings.Contains(param.Name, "Loopback") {
			netDeviceListS = append(netDeviceListS, param.Name)
		}
	}

	return netDeviceListS
}

func anyPcapAddress(these []pcap.InterfaceAddress, other net.IP) bool {
	if nil == other {
		return false
	}
	for _, cur := range these {
		if cur.IP.Equal(other) {
			return true
		}
	}
	return false
}

func sameIP(netIface net.Interface, pcapIface pcap.Interface) bool {
	addrs, err := netIface.Addrs()
	if err != nil {
		log.Fatal(err)
		return false
	}
	for _, addr := range addrs {
		addrNoPort, _, _ := net.ParseCIDR(addr.String())
		if anyPcapAddress(pcapIface.Addresses, addrNoPort) {
			return true
		}
	}
	return false
}

func getWindowsNetworkDeviceAddr(networkName string) string {

	result := ""

	netIfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	var guiSelectedInterface net.Interface
	for _, iface := range netIfaces {
		if iface.Name == networkName {
			guiSelectedInterface = iface
			break
		}
	}

	pcapIfaces, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, pcd := range pcapIfaces {
		if sameIP(guiSelectedInterface, pcd) {
			result = pcd.Name
			break
		}
	}

	return result
}

func capturePackets(networkDevice string, dstIP net.IP, dstPort int) {

	var snapshotLength int32 = 1024 // the maximum size to read for each packet
	var promiscuous bool = false    // interface in promiscuous mode

	// Open network device
	handle, err := pcap.OpenLive(getWindowsNetworkDeviceAddr(networkDevice), snapshotLength, promiscuous, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set pcap filter
	filter := "udp and port " + strconv.Itoa(dstPort) + " and ip broadcast"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set pcap filter to: " + filter + ".")

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// forward each caputred packet
		log.Println("Broadcast packet was captured and will be forwareded as unicast to " + dstIP.String())
		forwardPacket(dstIP, dstPort, packet)
	}
}

func forwardPacket(dstIP net.IP, dstPort int, packet gopacket.Packet) {

	serverAddr, err := net.ResolveUDPAddr("udp4", dstIP.String()+":"+strconv.Itoa(dstPort))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenPacket("udp", ":"+strconv.Itoa(dstPort))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// get payload from captured packet
	data := packet.ApplicationLayer().Payload()

	// send new unicast packet to server
	_, err = conn.WriteTo(data, serverAddr)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Packet was successfully forwarded as unicast to: " + dstIP.String())
	}
}
